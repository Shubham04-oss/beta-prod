# Google Cloud Agent Identity

Agent Identity provides a strongly attested, cryptographic identity for each agent that is based on the [SPIFFE standard](https://spiffe.io/). With Agent Identity, your agent can securely authenticate to MCP servers, cloud resources, endpoints, and other agents, acting either on its own behalf or on behalf of an end user.

Unlike service accounts, agent identities are not shared by multiple workloads by default, can't be impersonated, and don't allow developers to generate long-lived service account keys. Access tokens generated for Google Cloud are cryptographically bound to the agent's unique X.509 certificates to prevent token theft.

---

## Authentication Models

To authenticate with various tools and services, Agent Identity supports several authentication models. The model that an agent uses depends on the authentication method offered by the target resource and whether the agent acts on its own authority or on behalf of an end user.

| Authority | Authentication method | Target resource | Use case and solution |
|---|---|---|---|
| **User-delegated authority** | OAuth 2.0 (3-legged) | External tools and services | When an agent acts on behalf of a specific user (for example, to access a user's Jira tasks or GitHub repositories). You configure a 3-legged OAuth auth provider in Agent Identity auth manager to manage user consent and tokens. |
| **Agent's own authority** | Cloud-based identity (Agent Identity) | Google Cloud services | When an agent hosted on Google Cloud needs to access other Google Cloud services using its own identity. |
| **Agent's own authority** | OAuth 2.0 (2-legged) | External tools and services | Recommended for machine-to-machine authentication with external services that support OAuth. You configure a 2-legged OAuth auth provider in Agent Identity auth manager to handle client credentials and access tokens. |
| **Agent's own authority** | API key | External tools and services | For external services that require a cryptographic key or password for authentication. You configure an API key auth provider in Agent Identity auth manager to help securely store and manage the keys. |

---

## Core Components

Agent Identity involves several key components that together help provide secure authentication and authorization.

### 1. SPIFFE-based identity
Each agent is assigned a unique identity string, or SPIFFE ID, based on the SPIFFE standard. This identity is strongly attested, tied to the agent's lifecycle, and mapped directly to the resource URI where the agent is hosted.

The identity follows this format:
> `spiffe://TRUST_DOMAIN/resources/SERVICE/RESOURCE_PATH`

For example:
> `spiffe://agents.global.org-123456789012.system.id.goog/resources/aiplatform/projects/9876543210/locations/us-central1/reasoningEngines/my-test-agent`

### 2. Agent Credentials
Agent credentials provide cryptographic proof of an agent's identity. The system supports X.509 certificates and Google Cloud access tokens. An X.509 certificate is auto-provisioned and managed on the agent to help support stronger authentication.

### 3. Agent Identity Auth Manager
Agent Identity auth manager is a credential vault designed to help protect credentials. It lets agents authenticate using an API key or OAuth client ID and secret, or on behalf of a user through OAuth delegation using end-user access tokens. Within the auth manager, you configure auth providers that define the authentication type and credentials for specific third-party applications.

---

## Security & Governance

Agent Identity is fully integrated with Google's policy systems like IAM, Principal Access Boundary (PAB), and VPC Service Controls.

* **Context-Aware Access:** By default, a Google-managed Context-Aware Access policy secures Agent Identity credentials. Beyond the Agent Gateway, the policy enforces Demonstrable Proof of Possession (DPoP) by authenticating the agent's access token. The policy also enforces that mTLS is used to access the Agent Gateway.
* **IAM Integration:** Support for standard IAM allow policies and deny policies.
* **Principal Access Boundary (PAB):** A PAB limits the resources an agent can access, regardless of other permissions.
* **VPC Service Controls:** Support for using agent identities in ingress and egress rules to allow access to resources protected by a service perimeter.
