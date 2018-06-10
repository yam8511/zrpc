# Go-ZRPC

###### tags `rpc` `container` `golang` `framework`

---

# Feature

1. Using golang standard RPC/JSON-RPC library
2. Not depends on registry (Only need configuration)
3. HTTP to RPC / JSON to RPC (Not Support HTTP2)
4. Easy to use

---

# Example
> 待補

---

# TODO
- [x] RPC
- [ ] Proxy
- [ ] UI Web

---

# Motivation
After designing and developing some projects, that running in container, I reviewed my projects. since the feature of docker-compose's [networking](https://docs.docker.com/compose/networking/), or kubernetes's [services](https://kubernetes.io/docs/concepts/services-networking/service/) that is like load balance, I think that I don't need registry, like [etcd](https://coreos.com/etcd/) or [consul](https://consul.io). the only one which I need is a known service list in sidecar. So, I decided to create a framework that not need depends on registry, and I can list known  services by myself for sidecar to call services.
> [time=Thu, May 24, 2018 11:48 PM][name=Zuolar]
