# Security

Velora is an early foundation, not a hardened hosting control plane. Use the newest reviewed revision; no long-term supported release series exists yet.

Report vulnerabilities privately through the repository's GitHub “Report a vulnerability” facility when enabled. If unavailable, contact repository maintainers privately before sharing exploit details. Do not file public issues containing credentials or information about live deployments. Include affected revision, impact, reproduction and suggested mitigation where possible. No security contact address is invented by this scaffold.

Deployment requirements: HTTPS, Secure cookies, same-origin proxying, private PostgreSQL/Redis networking, restricted database credentials and backups. Only administrators create Station records. The initial CLI is a privileged local administrative surface; restrict host/container access. No default account is shipped.

Current limits: one Core replica, in-memory peer-IP login throttling, no MFA/password reset/session-management UI/audit trail. Proxy traffic shares a conservative limit; forwarding headers are not trusted. Sessions expire after 12 hours; expired rows can be periodically removed by an operator with `DELETE FROM sessions WHERE expires_at < now()`. Redis does not control authorization and failures do not disable security checks. Revisit these limits before multi-tenant public deployment.
