# Hooklay - webHook relay middleware

### config
```yaml
serverPort: 8080
hook: /dingtalk/alerts
security:
  token: 
    header: Authorization
    value: 123456
templates:
- name: mysqlaudit
  content: |
    {"type": "markdown", "msg": {"title", "{{.title}}", "content": "content1 {{.message}}"}}
- name: javalogerror
  content: |
    {"type": "markdown", "msg": {"title", "{{.title}}", "content": "content2 {{.message}}"}}
targets:
- name: mysqlaudit
  enabled: true
  url: https://api.dingtalk.com/robot?ak=23249304329435945
  bodyTemplate: mysqlaudit
  conditions:
  - key: $.alerts[0].labels.alert_type
    operator: Eq
    value: mysqlaudit
- name: javalogerror
  enabled: true
  url: https://api.dingtalk.com/robot?ak=23249304329435945
  bodyTemplate: javalogerror
  conditions:
  - key: $.alerts[0].labels.alert_type
    operator: Eq
    value: javalogerror
```