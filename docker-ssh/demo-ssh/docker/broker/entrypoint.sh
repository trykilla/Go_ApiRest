#!/bin/bash

# Configuración de iptables para permitir la comunicación en la red dmz

echo "10.0.1.4 myserver.local" >> /etc/hosts

ip route del default
ip route add default via 10.0.1.2 dev eth0
# Limpiar reglas existentes
# iptables -F
# iptables -P INPUT DROP
# iptables -P FORWARD DROP
# iptables -P OUTPUT ACCEPT

# # Permitir tráfico en interfaz loopback
# iptables -A INPUT -i lo -j ACCEPT
# iptables -A OUTPUT -o lo -j ACCEPT

# # Permitir ping
# iptables -A INPUT -p icmp -j ACCEPT
# iptables -A OUTPUT -p icmp -j ACCEPT



# # Permitir comunicación con otros contenedores en la red srv
# iptables -A INPUT -s 10.0.2.0/24 -j ACCEPT
# iptables -A OUTPUT -d 10.0.2.0/24 -j ACCEPT

# iptables -A INPUT -s 192.168.1.0/24 -j ACCEPT
# iptables -A OUTPUT -d 192.168.1.0/24 -j ACCEPT

# # Configurar reenvío de paquetes entre interfaces
# sysctl -w net.ipv4.ip_forward=1

# Iniciar tu aplicación o servicios aquí

# Ejecutar el comando proporcionado o iniciar el shell si no se proporciona ningún comando
if [ -z "$1" ]; then
    exec /bin/bash
else
    exec "$@"
fi
