# WebHook Relay

### config like this
```yaml
apiVersion: 1
serverPort: 8080
relays:
- hook: /hook1
  enabled: true
  security:
    token: 
      header: X-Gitlab-Token
      value: 123456
    signature:
      method: HMAC
      password: 123456
      header: X-Signature
  targets:
  - name: test
    enabled: true
    url: https://igo.pub
    body: |
      {"branch": "{{.branch}}"}
    conditions:
    - key: branch
      operator: Eq
      value: t2
  - name: android
    enabled: true
    url: https://igo.pub/signin
    body: |
      {"targetBranch": "{{.target.branch}}"}
    conditions:
    - key: target.branch
      operator: Eq
      value: t2
```