# Docker highway

Love is a highway, maybe docker should be as well?

This is an unstructered attempt at getting some ease of use into
docker service discovery of both local (dev) and clustered
(production) containers.

## Why would I want this?

Well, it's not certain that you do. This switches the usual linking
between docker containers for an off-brand etcd backed solutions.

This allows you to not worry about linking between containers, but
instead of \$linkname_PORT_xyz_TCP_ADDR:\$linkname_PORT_xyz_TCP_PORT 
can say container_name.local. It also enables you to continue this
practice for containers running on non-local machines.

## What is this?

It consists of several components:

- etcd
- Haproxy + confd
- dnsmasq

Combined, this allows us to store docker containers' ip+port address
in etcd, saved under the name of the container we can use haproxy to
match host headers and resolve the containers based on URLs. Cherry on
top is dnsmasq, which enables us to do all of this on the docker
daemon address (default 172.17.42.1). 

## What must I do to partake in this?!

Running this is the easy bit:

> make

This will boot 4 containers, assuming that ports 80 and 53 are
available on your system.

Next comes the DNS resolving. To enable intra-container communication
with this scheme you must use the --dns flag either on all containers
you start or on your docker daemon. The value of this should be
the ip of your docker network. This can be specified on your daemon,
default is 172.17.42.1, so:

> --dns 172.17.42.1

Now, in order for containers to
be resolvable bind the service you want to advertise to the docker
network ip. For instance for port 9090:

> -p 172.17.42.1::9090

To be able to use the same address to reach machines from your own
machine you have to have 127.0.0.1 in your list of dns nameservers.
Objective is to have this in /etc/resolv.conf, the road to achieve
this varies greatly between setups. Please consult your manual

## Current shortcomings

Only supports http services, because of the header based approach.
This can be extrapolated to known ports for tcp, but it would require
us to run everything with --net=host, which brings a series of side
effects.
