apiVersion: v1
kind: ConfigMap
metadata:
  name: pay-mysql-cm
data:
  dsn: "pay-gateway:$pass@tcp(pay-mysqlha-0.pay-mysqlha:3306)/pay_gateway?charset=utf8mb4&parseTime=true&loc=Local&timeout=10s&collation=utf8mb4_unicode_ci"
