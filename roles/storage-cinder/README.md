# Ansible Meta Role to Deploy Standalone Cinder

This role aggregates Cinder, RabbitMQ and MariaDB to deploy Standalone
Cinder with no authentication (noauth). 

MariaDB is deployed on a node with label=mariadb. MariaDB uses hostPath
for storage. 
