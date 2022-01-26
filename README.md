# linode-tools

A suite of tools to automate some parts of Linode.  


## Kube-Linode

A janky fix for dealing with maintaining IPTables rules for Debian/Ubuntu systems using UFW.

The reason for this tool includes keeping iptables rules synchronized with your Linode Kubernetes Cluster
to ensure only nodes from your cluster can communicate with services running on a Linode instance.  For example,
let's say you have mongodb running on a Linode and you only want your LKS nodes to communicated with MongoDB.  This
tool helps acheive that goal by monitoring LKS for nodes and building a list of IP Addresses to add to UFW.  

The basic premise is to generate the user.rules file and reload UFW with the new rules when changes
are detected

TODO:  Assumes SSH is open to the world. Maybe make an additional tool that tracks IP Address for allowing ssh, like from your home network.  

### Usage

kube-linode -rules /etc/ufw/user.rules -ufw /usr/sbin/ufw

