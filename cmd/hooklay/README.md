# Hooklay - webHook relay middleware
### Overview
a tools forward web hook to other web hook endpoint with json render
### Example Config
```yaml
serverPort: 8080
hook: /dingtalk/alerts
security:
  token:
    header: Authorization
    value: "Basic YWRtaW46MTIzMTIz"
templates:
- name: mysqlaudit
  content: |
    {
        "msgtype": "markdown",
        "markdown": {
          "title": "{{ (index .alerts 0).labels.alertname }}",
          "text": "# {{ (index .alerts 0).labels.alertname }} \n {{ range .alerts}} ___\n- user: {{.labels.user}}  \n  ip: {{.labels.ip}}  \n  time: {{.labels.date}}  \n  db: {{.labels.db_name}}  \n  sql: {{.labels.query}}  \n  {{ end }}"
        }
    }
- name: mysqlaudit-idempotent
  content: |
    {{ (index .alerts 0).labels.query }}-{{ (index .alerts 0).labels.date }}-{{ (index .alerts 0).labels.user }}
targets:
- name: mysqlaudit
  enabled: true
  url: https://oapi.dingtalk.com/robot/send?access_token=xxx
  bodyTemplate: mysqlaudit
  idempotentTemplate: mysqlaudit-idempotent
  idempotentDruationSeconds: 50
  conditions:
  - key: $.alerts[0].labels.alert_type
    operator: Eq
    value: mysqlaudit
```

Expose a http endpoint `/dingtalk/alerts`, use `basic auth` based authentication, and send http request to `https://oapi.dingtalk.com/robot/send?access_token=xxx` use `mysqlaudit` template