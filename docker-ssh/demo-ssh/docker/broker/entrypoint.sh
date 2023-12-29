#!/bin/bash

# Configuración de iptables para permitir la comunicación en la red dmz

echo "10.0.1.4 myserver.local" >> /etc/hosts

ip route del default
ip route add default via 10.0.1.2 dev eth0

# Limpiar reglas existentes
iptables -F
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

# Permitir tráfico proveniente del router (reemplaza x.x.x.x con la IP del router)
iptables -A INPUT -i eth0 -s 10.0.1.2 -j ACCEPT

# Permitir tráfico en el puerto 5000
iptables -A INPUT -i eth0 -s 10.0.1.2 -p tcp --sport 5000 -j ACCEPT
iptables -A INPUT -i eth0 -s 10.0.1.2 -p tcp --dport 5000 -j ACCEPT

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
iptables -A OUTPUT -p icmp -j ACCEPT

sysctl -w net.ipv4.ip_forward=1

# Iniciar tu aplicación o servicios aquí

# Ejecutar el comando proporcionado o iniciar el shell si no se proporciona ningún comando
./broker
