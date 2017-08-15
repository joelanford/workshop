Global Resource
```
    Deployment (nginx-ingress-controller)
```

Desk Resources
```
    Namespace (trusted)
        ServiceAccount (user)

        RoleBinding (trusted-user-view)

        Deployment (kubeshell)
        Service (kubeshell)
        Ingress (kubeshell)

    ResourceQuota (default)
    Namespace (default)
        RoleBinding (trusted-user-edit)

```

Workshop controller (1 per cluster)
  Desk CRD Controller (1 per workshop controller)
  Desk Controller (1 per desk resource)
    ServiceAccount (user)
    Trusted Namespace Controller (1 per desk controller)
        Namespace (trusted)
        RoleBinding (trusted-user-view)
        Deployment (kubeshell)
        Service (kubeshell)
        Ingress (kubeshell; if workshop domain is set)
    Untrusted Namespace Controller (1 per desk controller)
        Namespace (default)
        RoleBinding (trusted-user-edit)
        ResourceQuota (default)
