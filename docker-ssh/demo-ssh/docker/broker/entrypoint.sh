#!/bin/bash

# Modify /etc/hosts
echo "172.17.0.2   myserver.local" >> /etc/hosts


# Run your application
./broker
