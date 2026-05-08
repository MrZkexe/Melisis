If you're curious to see the development process behind this tool, [click here](hhttps://zksec.xyz/post/?postuuid=ad9872c5-1fd9-45a7-b3d5-1fcfe7a3286c).

*Note: this is written in Brazilian Portuguese.*
<center>
<img src="icone.png">
</center>

> Icon generated with AI. - **Temporary**
---
# What is Melisis?

**Melisis** is an open-source honeypot that simulates an SSH service in order to detect, deceive, and block access from potential malicious actors.

---

> The name **Melisis** is a wordplay involving **honey (mel)**, the song **Melissa** from **Fullmetal Alchemist**, and the name **Isis**, referencing the pianist **Isis Vasconcellos**, [reference](https://www.youtube.com/watch?v=VZ9tSbb1RiQ).
---

## Dependencies

To run the honeypot, you need:

- `fail2ban`
- `wget`
- `curl`
- A Linux system with **systemd**

> This project was developed with Debian-based systems in mind, but it works on other distributions with minor adjustments.

---

## Installation

### Debian / Ubuntu
```bash
sudo apt update
sudo apt install fail2ban wget curl -y
```
Then run the installer:
```bash
 curl -sL https://raw.githubusercontent.com/MrZkexe/Melisis/refs/heads/main/setup.sh | sudo bash
```
Or, if you prefer to clone manually:
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
> In some CentOS/RHEL versions, `fail2ban` is available via the EPEL repository.

Then:
```bash
curl -sL https://raw.githubusercontent.com/MrZkexe/Melisis/refs/heads/main/setup.sh | sudo bash
```
Or:
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
Then:
```bash
curl -sL https://raw.githubusercontent.com/MrZkexe/Melisis/refs/heads/main/setup.sh | sudo bash
```
Or:
```bash
git clone https://github.com/MrZkexe/Melisis.git  
cd pasta      
chmod +x setup.sh      
sudo ./setup.sh
```
---

## Enabling the Service

After installation, you must enable and start the Melisis service using **systemd**:
```bash
sudo systemctl enable melisis    
sudo systemctl start melisis
```
### Check status
```bash
sudo systemctl status melisis
```
---

### Restart the service (after configuration changes)
```bash
sudo systemctl restart melisis
```
---

### Stop the service
```bash
sudo systemctl stop melisis
```
---

## Post-installation

After installation, the following files will be created:

- `/etc/Melisis/melisis.conf` → **Honeypot configuration**
- `/var/log/Melisis.txt` → **Block logs**
- `/var/log/MelisisCommandsLog.txt` -> **Command log from the fake shell**
- `/etc/fail2ban/jail.d/melisis.conf` → **Ban rules**
- `/etc/fail2ban/filter.d/melisis.conf` → **Filter used by fail2ban**

> The most important files for the user are the **configuration** and **ban rules**.

---

## Melisis Configuration

### File: `/etc/Melisis/melisis.conf`
```yaml
[conf]  
# Set the IP address of the network interface where you want to run the honeypot  
ip = 0.0.0.0  
  
# Port where the honeypot will run; default is 22 since it simulates SSH  
port = 22  
  
# Modes range from 0 to 3 with the following behaviors:  
# mode 0: extreme mode; bans the IP at any sign of connection  
# mode 1: deceives scanners by behaving like a legitimate SSH server  
# mode 2: allows access to a fake shell and blocks the IP after disconnection  
# mode 3: similar to mode 2, but does not ban and uses social engineering  
mode = 0
```
### Parameters

#### `ip`

Defines the IP address of the interface where the honeypot will run.

- `0.0.0.0` → listens on all interfaces (recommended)
- `127.0.0.1` → local only (for testing)
- specific `IP` → listens only on that interface

---

#### `port`

> Defines the port where the honeypot will run.

- `22` (default) → more realistic (simulates real SSH)
- Other ports can be used for testing

> **Warning:** if you already use SSH on port 22, you must change one of them.

---

#### `mode`

Defines the honeypot behavior.

##### Mode 0 — Aggressive

- Immediately bans the IP upon any connection attempt
- Fully focused on protection
- Collects minimal data

---

##### Mode 1 — Scanner deception

- Simulates a legitimate SSH server
- Tricks automated tools (e.g., nmap, bots)
- Blocks after suspicious behavior

---

##### Mode 2 — Interactive with blocking

- Allows access to a fake shell
- Bans after disconnection

> Best balance between security and data collection

---

##### Mode 3 — Social engineering

- Similar to mode 2, with access to a fake shell
- Does **not** ban the IP
- Whenever a command is executed, returns an error indicating a missing 'SSH plugin'
- Then suggests a (simulated) installation command

> Goal: trick the attacker into executing commands outside the honeypot environment  
> Use with caution — depends on attacker behavior and does not guarantee results

---

## Fail2Ban Configuration

### File: `/etc/fail2ban/jail.d/melisis.conf`
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

### Explanation

- `enabled` → enables monitoring
- `filter` → defines the filter used
- `logpath` → log file being analyzed
- `maxretry` → number of attempts before banning
- `findtime` → time window for counting attempts
- `bantime` → duration of the ban
- `action` → type of block (all ports)

---

## Monitoring

Follow logs in real time:

Banned IPs

```
tail -f /var/log/Melisis.txt
```

View currently banned IPs

```
sudo fail2ban-client status melisis
```

View command attempts

```
tail -f /var/log/MelisisCommandsLog.txt
```

All logs in real time

```
sudo journalctl -u melisis -n 50 -f
```

---

## Final considerations

Melisis is a simple and effective tool to:

- Detect automated attacks
- Deceive scanners and bots
- Protect real services
