# Private networking

Literal OVH VPS products use a WireGuard overlay because they do not expose the
same project-private-network attachment as OVH Public Cloud instances.

- Dokploy: `10.20.0.10`
- Go: `10.20.0.20`
- ELT: `10.20.0.30`
- future database: `10.20.0.40`

Terraform's OVH edge firewalls restrict SSH, allow WireGuard only between known
peer public IPs, expose 80/443 only on Dokploy and Go, and deny all remaining
public IPv4 ingress. Host UFW rules use the selected environment's WireGuard CIDR
for private services. ELT gRPC, PostgreSQL, PGMQ, PgBouncer, and
Dokploy-to-target SSH use private addresses.

The `simulation-shared` Terraform module creates the OVH Public Cloud region,
private subnet, and SSH key used by dynamic hosts. The Go manager still owns each
temporary host. The literal Go VPS is not automatically attached to that subnet.
Before enabling simulations, provide an OVH-supported router or a WireGuard peer
in the dedicated simulation image and prove private Go-to-host connectivity. The
current Microsandbox cloud-init does not create this cross-network route. Public
Microsandbox or runtime control ports are never an acceptable fallback.

Dynamic SSH uses strict host verification. Persistent-host bootstrap consumes
operator-verified `known_hosts` data rather than learning trust with `ssh-keyscan`.
Dynamic hosts present a certificate signed by the dedicated simulation host CA.
See [setup-guide.md](../setup/ovh-dokploy-microsandbox.md) for the complete routing and trust checks.
