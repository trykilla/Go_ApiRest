#!/bin/bash

# Configuración de iptables para permitir la comunicación en la red srv
ip route del default
ip route add default via 10.0.2.2 dev eth0

# Limpiar reglas existentes
iptables -F
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

sysctl -w net.ipv4.ip_forward=1


iptables -A INPUT -i eth0 -s 10.0.1.4 -j ACCEPT # puerto que se comunica con el broker 10.0.2.3:8084

# permito tráfico en el 8084

iptables -A INPUT -i eth0 -s 10.0.1.4 -p tcp --sport 8084 -j ACCEPT
iptables -A INPUT -i eth0 -s 10.0.1.4 -p tcp --dport 8084 -j ACCEPT

# Permitir respuestas DNS
iptables -A INPUT -i eth0 -p udp --sport 53 -j ACCEPT
iptables -A INPUT -i eth0 -p tcp --sport 53 -j ACCEPT

# Permitir respuestas HTTP
iptables -A INPUT -i eth0 -p tcp --sport 80 -j ACCEPT

# Permitir respuestas HTTPS
iptables -A INPUT -i eth0 -p tcp --sport 443 -j ACCEPT
# Permitir tráfico en interfaz loopback
iptables -A INPUT -i lo -j ACCEPT

# Permitir ping
iptables -A INPUT -p icmp -j ACCEPT

iptables -A INPUT -p tcp --dport 22 -s 10.0.2.2 -j ACCEPT
iptables -A INPUT -p tcp --dport 22 -s 10.0.3.3 -j ACCEPT
iptables -A INPUT -p tcp --sport 22 -s 10.0.3.0/24 -j ACCEPT

echo -e "Match Address 10.0.3.3\n  AllowUsers op\n" >>/etc/ssh/sshd_config


service ssh start
service rsyslog start
./auth
