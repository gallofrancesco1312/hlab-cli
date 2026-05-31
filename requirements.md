# hlab — Requisiti di progetto

> CLI in Go per la gestione remota dell'homelab tramite agenti mTLS e discovery multicast

---

## 1. Visione generale

`hlab` è una CLI (client) + daemon (agente) per controllare i servizi del proprio homelab da qualsiasi macchina sulla rete locale. L'utente può scoprire automaticamente i nodi disponibili, interrogarli sui servizi gestiti, e inviare comandi (start, stop, restart, status) senza configurare nulla manualmente per ogni nuovo nodo.

L'ispirazione è `kubectl`, ma orientata a un singolo operatore su una LAN casalinga con più macchine.

---

## 2. Componenti

### 2.1 `hlab` — CLI (lato workstation)

Binario Go installato sul computer dell'utente. Responsabile di:

- Eseguire i comandi dell'utente (`hlab nas jellyfin stop`)
- Scoprire i nodi via multicast
- Mantenere una cache locale dei nodi noti (`~/.hlab/nodes.json`)
- Autenticarsi verso gli agenti tramite mTLS
- Presentare output leggibile (tabelle, stati, log)

### 2.2 `hlab-agent` — Daemon (lato server)

Binario Go da deployare su ogni macchina dell'homelab. Responsabile di:

- Ascoltare su una porta HTTP(S) dedicata (default `:8443`)
- Rispondere ai comandi inviati dalla CLI
- Eseguire operazioni su Docker (standalone + Compose) e systemd
- Annunciare la propria presenza via multicast UDP

---

## 3. Discovery multicast

### Meccanismo

Ogni agente invia periodicamente un beacon UDP su un indirizzo multicast (es. `239.255.42.42:9999`) sulla LAN locale. Il beacon contiene solo le informazioni minime di presenza.

### Payload del beacon (JSON)

```json
{
  "node":    "homelab-nas",
  "addr":    "192.168.1.50",
  "port":    8443,
  "version": "0.1.0"
}
```

Il beacon **non** include la lista dei servizi — quella viene recuperata dalla CLI con una chiamata REST separata dopo la discovery.

### Comportamento CLI

- `hlab discover` — ascolta per N secondi (default 3s) e stampa i nodi trovati
- I nodi vengono salvati in `~/.hlab/nodes.json` con timestamp dell'ultimo avviso
- Un nodo non visto da più di X minuti viene marcato come `stale` (configurabile)
- La CLI può operare anche senza discovery se il nodo è già in cache

### Frequenza beacon

- Default: ogni 10 secondi
- Configurabile via flag dell'agente: `--beacon-interval 30s`

---

## 4. Autenticazione — mTLS

Tutta la comunicazione CLI ↔ agente avviene su TLS mutuamente autenticato.

### PKI

Una CA self-signed gestita dall'utente. Struttura file consigliata:

```
~/.hlab/
  ca.crt           # CA pubblica condivisa
  client.crt       # Certificato client (firmato dalla CA)
  client.key       # Chiave privata client
  nodes.json       # Cache dei nodi scoperti
```

Sul server (`/etc/hlab/`):
```
  ca.crt           # Stessa CA
  server.crt       # Certificato server (firmato dalla CA)
  server.key       # Chiave privata server
```

### Flusso

1. La CLI apre connessione TLS e presenta `client.crt`
2. L'agente verifica che il certificato sia firmato dalla CA di fiducia
3. L'agente presenta `server.crt`
4. La CLI verifica che il certificato sia firmato dalla stessa CA
5. Solo se entrambe le verifiche passano la connessione è stabilita

### Generazione certificati

`hlab pki init` — comando CLI per generare CA + certificati client/server con defaults ragionevoli (2048-bit RSA o curve25519, validità 3 anni).

---

## 5. API dell'agente

L'agente espone una REST API minimale su HTTPS.

### Endpoints

| Metodo | Path | Descrizione |
|--------|------|-------------|
| `GET`  | `/health` | Heartbeat, versione agente |
| `GET`  | `/services` | Lista tutti i servizi gestiti e il loro stato |
| `POST` | `/services/{name}/start` | Avvia un servizio |
| `POST` | `/services/{name}/stop` | Ferma un servizio |
| `POST` | `/services/{name}/restart` | Riavvia un servizio |
| `GET`  | `/services/{name}/logs` | Ultimi N log (query param `?lines=50`) |

### Risposta `/services`

```json
{
  "node": "homelab-nas",
  "services": [
    {
      "name":    "jellyfin",
      "type":    "compose",
      "status":  "running",
      "uptime":  "3d 4h 12m"
    },
    {
      "name":    "traefik",
      "type":    "compose",
      "status":  "running",
      "uptime":  "7d 1h 00m"
    },
    {
      "name":    "ssh",
      "type":    "systemd",
      "status":  "running",
      "uptime":  "14d"
    }
  ]
}
```

---

## 6. Runtime supportati

L'agente supporta tre tipi di backend per l'esecuzione dei comandi:

### 6.1 Docker Compose

- Rilevamento automatico: scansiona directory configurate (default `/opt/stacks/`, configurabile) cercando `docker-compose.yml` o `compose.yaml`
- Comandi eseguiti: `docker compose -f <path> up -d`, `down`, `restart`, `logs`
- Il nome del servizio esposto all'utente è il nome della directory

### 6.2 Docker standalone

- Rilevamento automatico: lista i container attivi e stopped via Docker socket (`/var/run/docker.sock`)
- Comandi eseguiti via API Docker (libreria Go `docker/client`), non shell
- Il nome del servizio è il nome del container

### 6.3 systemd

- Rilevamento automatico: lista le unit attive di tipo `service` via D-Bus o `systemctl list-units`
- Comandi eseguiti: `systemctl start|stop|restart <unit>` (richiede privilegi)
- Configurabile: whitelist delle unit esponibili (per sicurezza)

### Priorità in caso di conflitto di nomi

Se un container Docker e una unit systemd hanno lo stesso nome, l'agente applica la priorità: `compose > docker > systemd`. Configurabile.

---

## 7. Configurazione dell'agente

File YAML su ogni server (`/etc/hlab/agent.yaml`):

```yaml
node_name: homelab-nas        # Nome mostrato dalla CLI (default: hostname)
port: 8443
tls:
  ca_cert:     /etc/hlab/ca.crt
  server_cert: /etc/hlab/server.crt
  server_key:  /etc/hlab/server.key

beacon:
  enabled:  true
  interval: 10s
  multicast_addr: "239.255.42.42:9999"

backends:
  compose:
    enabled: true
    scan_dirs:
      - /opt/stacks
      - /home/user/compose
  docker:
    enabled: true
    socket: /var/run/docker.sock
  systemd:
    enabled: true
    unit_whitelist:
      - sshd
      - nginx
      - postgresql
```

---

## 8. Comandi CLI

### Sintassi generale

```
hlab [global flags] <nodo> <servizio> <azione> [flags]
```

### Comandi principali

```bash
# Discovery
hlab discover                        # Scopre nodi sulla LAN (3s timeout)
hlab discover --timeout 10s          # Timeout personalizzato

# Nodi
hlab nodes                           # Lista nodi noti in cache
hlab nodes --refresh                 # Forza re-discovery prima di listare

# Servizi (su un nodo specifico)
hlab nas services                    # Lista tutti i servizi su homelab-nas
hlab nas jellyfin status             # Stato di un servizio
hlab nas jellyfin start
hlab nas jellyfin stop
hlab nas jellyfin restart
hlab nas jellyfin logs               # Ultimi 50 log
hlab nas jellyfin logs --lines 200

# Tutti i nodi (broadcast)
hlab all services                    # Lista servizi su tutti i nodi noti
hlab all jellyfin status             # Status di jellyfin su ogni nodo che lo ha

# PKI
hlab pki init                        # Genera CA, client cert, server cert
hlab pki status                      # Mostra scadenze certificati

# Configurazione
hlab config                          # Mostra configurazione corrente
```

### Alias nodi

La CLI supporta alias per i nodi in `~/.hlab/config.yaml`:

```yaml
aliases:
  nas: homelab-nas
  pc:  homelab-pc
```

Così `hlab nas jellyfin stop` risolve `nas` → `homelab-nas` prima della lookup in cache.

---

## 9. Output CLI

La CLI produce output in due modalità:

- **Human** (default): tabelle con colori, icone stato (✓ ✗ ↻), uptime leggibile
- **JSON** (`--output json`): per scripting e integrazione con altri tool

Esempio output `hlab nas services`:

```
NODE: homelab-nas (192.168.1.50)
┌─────────────┬─────────┬──────────┬──────────────┐
│ SERVICE     │ TYPE    │ STATUS   │ UPTIME       │
├─────────────┼─────────┼──────────┼──────────────┤
│ jellyfin    │ compose │ ✓ running│ 3d 4h 12m    │
│ traefik     │ compose │ ✓ running│ 7d 1h 00m    │
│ ssh         │ systemd │ ✓ running│ 14d          │
│ transmission│ docker  │ ✗ stopped│ —            │
└─────────────┴─────────┴──────────┴──────────────┘
```

---

## 10. Struttura del repository

```
hlab/
├── cmd/
│   ├── hlab/          # Entry point CLI
│   └── hlab-agent/    # Entry point daemon
├── internal/
│   ├── agent/         # Logica REST server
│   ├── cli/           # Comandi cobra
│   ├── discovery/     # Multicast beacon + listener
│   ├── backends/
│   │   ├── compose/
│   │   ├── docker/
│   │   └── systemd/
│   ├── mtls/          # Setup TLS, PKI helper
│   └── config/        # Parsing config
├── pkg/
│   └── hlabapi/       # Tipi condivisi CLI ↔ agente
├── Makefile
└── README.md
```

---

## 11. Dipendenze Go principali

| Libreria | Uso |
|----------|-----|
| `cobra` | Framework CLI (comandi, flag, help) |
| `golang.org/x/net/ipv4` | Multicast UDP |
| `net/http` + `crypto/tls` | Server/client mTLS |
| `docker/client` | Docker API (senza shell) |
| `coreos/go-systemd/dbus` | D-Bus per systemd |
| `olekukonko/tablewriter` | Tabelle CLI |
| `gopkg.in/yaml.v3` | Parsing config YAML |

---

## 12. Fasi di sviluppo consigliate

### Fase 1 — Scheletro
- Struttura repo, Makefile, `cobra` base
- Config YAML + parsing
- Logging strutturato

### Fase 2 — Agente minimal
- HTTP server con mTLS
- Backend Docker Compose (solo start/stop/status)
- Endpoint `/health` e `/services`

### Fase 3 — CLI → Agente
- Client mTLS
- Comandi `services`, `start`, `stop`, `status`
- Output tabellare

### Fase 4 — Discovery
- Beacon multicast dall'agente
- Listener multicast nella CLI
- Cache nodi + `hlab discover`

### Fase 5 — Backend completi
- Docker standalone via API
- systemd via D-Bus
- Whitelist e priorità

### Fase 6 — PKI helper + polish
- `hlab pki init`
- Output JSON
- Alias nodi
- Comando `all`

---

## 13. Non-obiettivi (out of scope v1)

- UI web
- Gestione multi-utente / RBAC
- Metriche / grafici
- Notifiche push
- Supporto per reti WAN / VPN (la discovery multicast è LAN-only by design)
