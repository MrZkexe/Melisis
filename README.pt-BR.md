Se tiver curiosidade para ver o processo de criação dessa ferramenta [click aqui.](https://zksec.xyz/post/?postuuid=ad9872c5-1fd9-45a7-b3d5-1fcfe7a3286c)
<center>
<img src="icone.png">
</center>

> Ícone gerado com IA. - **Provisório**
---
# O que é o Melisis?

O **Melisis** é um honeypot open source que simula um serviço SSH com o objetivo de detectar, enganar e bloquear acessos de possíveis agentes maliciosos.

---

> O nome **Melisis** é um trocadilho com **mel**, a música **Melissa** de **Fullmetal Alchemist** e o nome **Isis**, em referência à Pianista **Isis Vasconcellos**, [referencia](https://www.youtube.com/watch?v=VZ9tSbb1RiQ).
---
## Dependências

Para rodar o honeypot, você precisa de:

- `fail2ban`
- `wget`
- `curl`
- Um sistema Linux com **systemd**

> O projeto foi desenvolvido com foco em sistemas baseados em Debian, mas funciona em outras distribuições com pequenas adaptações.

---

## Instalação

### Debian / Ubuntu
```bash
sudo apt update    
sudo apt install fail2ban wget curl -y
```
Depois, execute o instalador:
```bash
curl -sL https://raw.githubusercontent.com/MrZkexe/Melisis/refs/heads/main/setup.sh | sudo bash
```
Ou, se preferir clonar manualmente:
```bash
git clone https://github.com/MrZkexe/Melisis.git
cd pasta    
chmod +x setup.sh    
sudo ./setup.sh
```
---

### Red Hat / CentOS / Fedora
```bash
sudo dnf install epel-release -y    
sudo dnf install fail2ban wget curl -y
```
> Em algumas versões do CentOS/RHEL, o `fail2ban` está no repositório EPEL.

Depois:
```bash
curl -sL https://raw.githubusercontent.com/MrZkexe/Melisis/refs/heads/main/setup.sh | sudo bash
```
Ou:
```bash
git clone https://github.com/MrZkexe/Melisis.git
cd pasta    
chmod +x setup.sh    
sudo ./setup.sh
```
---

### Arch Linux
```bash
sudo pacman -Syu fail2ban wget curl
```
Depois:
```bash
curl -sL https://raw.githubusercontent.com/MrZkexe/Melisis/refs/heads/main/setup.sh | sudo bash
```
Ou:
```bash
git clone https://github.com/MrZkexe/Melisis.git
cd pasta    
chmod +x setup.sh    
sudo ./setup.sh
```
---

## Ativando o serviço

Após a instalação, é necessário habilitar e iniciar o serviço do Melisis no **systemd**:
```bash
sudo systemctl enable melisis  
sudo systemctl start melisis
```
### Verificar status
```bash
sudo systemctl status melisis
```
---

### Reiniciar o serviço (após alterar configurações)
```bash
sudo systemctl restart melisis
```
---

### Parar o serviço
```bash
sudo systemctl stop melisis
```
---

## Pós-instalação

Após a instalação, os seguintes arquivos serão criados:

- `/etc/Melisis/melisis.conf` → **Configuração do honeypot**
- `/var/log/Melisis.txt` → **Log de bloqueios**
- `/var/log/MelisisCommandsLog.txt` -> **Log dos comandos na shellfake**
- `/etc/fail2ban/jail.d/melisis.conf` → **Regras de bloqueio**
- `/etc/fail2ban/filter.d/melisis.conf` → **Filtro utilizado pelo fail2ban**

> Os arquivos mais importantes para o usuário são o de **configuração** e o de **regras de bloqueio**.

---

## Configuração do Melisis

### Arquivo: `/etc/Melisis/melisis.conf`
```yaml
[conf]
# Set the IP address of the network interface where you want to run the honeypot>
ip = 0.0.0.0

# Port where the honeypot will run; default is 22 since it simulates SSH and is >
port = 22

# Modes range from 0 to 3 with the following behaviors:
# mode 0: extreme mode; bans the IP at any sign of connection to the honeypot po>
# mode 1: deceives scanners by behaving like a legitimate SSH server, but blocks>
# mode 2: allows access to a fake shell and blocks the IP after disconnection
# mode 3: similar to mode 2, but does not ban; attempts to use social engineerin>
mode = 0
```

### Parâmetros

#### `ip`

Define o endereço IP da interface onde o honeypot será executado.

- `0.0.0.0` → escuta em todas as interfaces (recomendado)
- `127.0.0.1` → apenas local (para testes)
- `IP` específico → escuta apenas naquela interface

---

#### `port`
> Define a porta onde o honeypot será executado.
- `22` (padrão) → mais realista (simula SSH real)
- Outras portas podem ser usadas para testes

> **Atenção:** se você já utiliza SSH na porta 22, será necessário alterar uma das portas.

---
#### `mode`

Define o comportamento do honeypot.

##### Mode 0 — Agressivo

- Bloqueia o IP imediatamente ao detectar qualquer conexão
- Foco total em proteção
- Não coleta muitas informações

---

##### Mode 1 — Engano de scanners

- Simula um servidor SSH legítimo
- Engana ferramentas automatizadas (ex: nmap, bots)
- Bloqueia após comportamento suspeito

---

##### Mode 2 — Interativo com bloqueio

- Permite acesso a um shell falso
- Bloqueia após desconexão

> Melhor equilíbrio entre segurança e coleta de dados

---

#####  Mode 3 — Engenharia social
- Similar ao modo 2, com acesso a um shell falso
- Não realiza bloqueio de IP
- Sempre que um comando é executado, retorna um erro indicando ausência de um 'plugin SSH'
- Em seguida, sugere um comando de instalação (simulado)

> Objetivo: induzir o atacante a executar comandos fora do ambiente do honeypot
> Use com cautela — depende do comportamento do atacante e não garante resultados

## Configuração do Fail2Ban
### Arquivo: `/etc/fail2ban/jail.d/melisis.conf`
```yaml
[melisis]  
enabled = true  
filter = melisis  
logpath = /var/log/Melisis.txt  
maxretry = 1  
findtime = 5m  
bantime = 5d  
action = iptables-allports[name=MelisisHoneypot]
```
---

### Explicação

- `enabled` → ativa o monitoramento
- `filter` → define o filtro utilizado
- `logpath` → arquivo de log analisado
- `maxretry` → número de tentativas antes do ban
- `findtime` → intervalo de tempo para contagem
- `bantime` → tempo de bloqueio
- `action` → tipo de bloqueio (todas as portas)

---

## Monitoramento

Acompanhar logs em tempo real:
IPs que foram banidos
```bash
tail -f /var/log/Melisis.txt
```
Ver IPs banidos atuais:
```bash
sudo fail2ban-client status melisis
```
Ver tentativas de comandos
```bash
tail -f /var/log/MelisisCommandsLog.txt
```
Todos o logs em tempo real
```bash
sudo journalctl -u melisis -n 50 -f
```

---

> Considerações finais.
> O Melisis é uma ferramenta simples e eficiente para:
- Detectar ataques automatizados
- Enganar scanners e bots
- Proteger serviços reais
