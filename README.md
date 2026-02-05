## Shop Boundary

> [!NOTE]
> The Shop Boundary in this software system is primarily concerned with the management of goods and services.

| Service            | Description        | Language/Framework   | Docs                                   | Status                                                                                                                                                                        |
|--------------------|--------------------|----------------------|----------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| admin              | Shop admin         | Python (Django)      | [docs](./admin/README.md)              | [![App Status](https://argo.shortlink.best/api/badge?name=shortlink-admin&revision=true)](https://argo.shortlink.best/applications/shortlink-admin)                           |
| bff                | API Gateway        | NodeJS (Wundergraph) | [docs](./gateway/README.md)            | [![App Status](https://argo.shortlink.best/api/badge?name=shortlink-shop-gateway&revision=true)](https://argo.shortlink.best/applications/shortlink-shop-gateway)             |
| email-subscription | Email subscription | Python (Django)      | [docs](./email-subscription/README.md) | [![App Status](https://argo.shortlink.best/api/badge?name=shortlink-email-subscription&revision=true)](https://argo.shortlink.best/applications/shortlink-email-subscription) |
| oms                | Order management   | Temporal             | [docs](./oms/README.md)                | [![App Status](https://argo.shortlink.best/api/badge?name=shortlink-oms&revision=true)](https://argo.shortlink.best/applications/shortlink-oms)                               |
| oms-graphql        | GraphQL API Bridge | Coming soon          | [docs](./oms-graphql/README.md)        | [![App Status](https://argo.shortlink.best/api/badge?name=shortlink-oms-graphql&revision=true)](https://argo.shortlink.best/applications/shortlink-oms-graphql)               |
| pricer             | Price service      | Go                   | [docs](./pricer/README.md)             | [![App Status](https://argo.shortlink.best/api/badge?name=shortlink-pricer&revision=true)](https://argo.shortlink.best/applications/shortlink-pricer)                         |
| ui                 | Shop service       | JS/NextJS            | [docs](./ui/README.md)                 | [![App Status](https://argo.shortlink.best/api/badge?name=shortlink-shop-ui&revision=true)](https://argo.shortlink.best/applications/shortlink-shop-ui)                       |
| support     | Support service     | PHP                | [docs](./boundaries/delivery/support/README.md)     | [![App Status](https://argo.shortlink.best/api/badge?name=shortlink-support&revision=true)](https://argo.shortlink.best/applications/shortlink-support)         |
| geolocation | Geolocation service | Go                 | [docs](./boundaries/delivery/geolocation/README.md) | [![App Status](https://argo.shortlink.best/api/badge?name=shortlink-geolocation&revision=true)](https://argo.shortlink.best/applications/shortlink-geolocation) |
| courier-emulation | Courier emulation service | Go | [docs](./courier-emulation/README.md) | Coming soon |

### Docs

- [GLOSSARY.md](./GLOSSARY.md) - Ubiquitous Language of the Shop Boundary
- [README.md](./docs/ADR/README.md) - Architecture Decision Records
- [Services Plan](./docs/services-plan.md) - План функциональности по сервисам
