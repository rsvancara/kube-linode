# linode-tools

A suite of tools to automate some parts of Linode.  


## Kube-Mongo

A janky fix for dealing with maintaining IPTables rules for Debian/Ubuntu systems using IPTables for MongoDB.

The reason for this tool includes keeping iptables rules synchronized with your Linode Kubernetes Cluster
to ensure only nodes from your cluster can communicate with services running on a Linode instance.  For example,
let's say you have mongodb running on a Linode and you only want your LKS nodes to communicated with MongoDB.  This
tool helps achieve that goal by monitoring LKS for nodes and building a list of IP Addresses to add to IPTables chain.

The tool works by creating an IPTables chain called mongodb and appends all the rules to this chain.  This chain is attached to the 
filter table.  By creating a chain, you reduce the risk of conflicting with UFW or other firewall management system.  

### Usage
```bash
./kube-linode 
```

## Kube-Nginx

Polls Linode Kubernetes Service for changes in nodes and updates the node list in an upstream.conf file that you can use in a
Nginx configuration.  Once the configuration is set, it will call sysctemctl reload nginx to bring in all the changes.  

### usage 
```bash
./kube-nginx -config /path/to/upstream.conf -systemctl /path/to/systemctl
```

