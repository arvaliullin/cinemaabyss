## Задание 1. Проектирование архитектуры

**Диаграмма контейнеров C4 целевой (To-Be) архитектуры**

![Диаграмма контейнеров целевой архитектуры](schemas/images/target-system-containers.png)
[Исходный код диаграммы](schemas/target-system-containers.puml)


## Задание 2. Proxy и Kafka

**Запрос к API Gateway: GET /api/movies**

![Запрос к API Gateway](docs/images/proxy-api-movies.png)

**Результаты Postman-тестов**

![Результаты Postman-тестов](docs/images/tests-postman.png)

**Состояние топиков Kafka**

![Состояние топиков Kafka](docs/images/ui-kafka.png)


## Задание 3

**Вызов https://cinemaabyss.example.com/api/movies**

![Список фильмов через Ingress](docs/images/kubernetes-api-movies.png)

**Логи events-service после тестов**

```bash
kubectl -n cinemaabyss logs -l app=events-service
```

```text
{"time":"2026-09-28T20:26:52.805152471Z","level":"WARN","msg":"kafka connect attempt failed","service":"events","client":"producer","attempt":3,"max_attempts":30,"error":"kafka: client has run out of available brokers to talk to: dial tcp 10.96.134.30:9092: connect: connection refused"}
{"time":"2026-09-28T20:26:56.571603667Z","level":"WARN","msg":"kafka connect attempt failed","service":"events","client":"producer","attempt":4,"max_attempts":30,"error":"kafka: client has run out of available brokers to talk to: dial tcp 10.96.134.30:9092: connect: connection refused"}
{"time":"2026-09-28T20:27:00.600244293Z","level":"INFO","msg":"consumer started","service":"events","group":"events-service","topics":["movie-events","user-events","payment-events"]}
{"time":"2026-09-28T20:27:00.60295746Z","level":"INFO","msg":"events service listening","service":"events","address":":8082"}
{"time":"2026-09-28T20:39:12.981844633Z","level":"INFO","msg":"event published","service":"events","event_id":"movie-6-viewed","event_type":"movie","topic":"movie-events","partition":0,"offset":0}
{"time":"2026-09-28T20:39:12.990400674Z","level":"INFO","msg":"event consumed","service":"events","topic":"movie-events","partition":0,"offset":0,"event_id":"movie-6-viewed","event_type":"movie","payload":{"action":"viewed","movie_id":6,"title":"Test Movie Event","user_id":4}}
{"time":"2026-09-28T20:39:13.107879216Z","level":"INFO","msg":"event published","service":"events","event_id":"user-4-logged_in","event_type":"user","topic":"user-events","partition":0,"offset":0}
{"time":"2026-09-28T20:39:13.110806091Z","level":"INFO","msg":"event consumed","service":"events","topic":"user-events","partition":0,"offset":0,"event_id":"user-4-logged_in","event_type":"user","payload":{"action":"logged_in","timestamp":"2026-09-28T20:39:13.098Z","user_id":4,"username":"testuser"}}
{"time":"2026-09-28T20:39:13.231337174Z","level":"INFO","msg":"event published","service":"events","event_id":"payment-4-completed","event_type":"payment","topic":"payment-events","partition":0,"offset":0}
{"time":"2026-09-28T20:39:13.234836341Z","level":"INFO","msg":"event consumed","service":"events","topic":"payment-events","partition":0,"offset":0,"event_id":"payment-4-completed","event_type":"payment","payload":{"amount":9.99,"method_type":"credit_card","payment_id":4,"status":"completed","timestamp":"2026-09-28T20:39:13.222Z","user_id":4}}
```

## Задание 4

**Развертывание Helm**

```bash
helm install cinemaabyss ./src/kubernetes/helm --namespace cinemaabyss --create-namespace
```

![Установка Helm](docs/images/helm-deploy.png)

**Вызов https://cinemaabyss.example.com/api/movies**

![Список фильмов после установки Helm](docs/images/helm-api-movies.png)


# Задание 5
Компания планирует активно развиваться и для повышения надежности, безопасности, реализации сетевых паттернов типа Circuit Breaker и канареечного деплоя вам как архитектору необходимо развернуть istio и настроить circuit breaker для monolith и movies сервисов.

```bash

helm repo add istio https://istio-release.storage.googleapis.com/charts
helm repo update

helm install istio-base istio/base -n istio-system --set defaultRevision=default --create-namespace
helm install istio-ingressgateway istio/gateway -n istio-system
helm install istiod istio/istiod -n istio-system --wait

helm install cinemaabyss .\src\kubernetes\helm --namespace cinemaabyss --create-namespace

kubectl label namespace cinemaabyss istio-injection=enabled --overwrite

kubectl get namespace -L istio-injection

kubectl apply -f .\src\kubernetes\circuit-breaker-config.yaml -n cinemaabyss

```

Тестирование

# fortio
```bash
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.25/samples/httpbin/sample-client/fortio-deploy.yaml -n cinemaabyss
```

# Get the fortio pod name
```bash
FORTIO_POD=$(kubectl get pod -n cinemaabyss | grep fortio | awk '{print $1}')

kubectl exec -n cinemaabyss $FORTIO_POD -c fortio -- fortio load -c 50 -qps 0 -n 500 -loglevel Warning http://movies-service:8081/api/movies
```
Например,

```bash
kubectl exec -n cinemaabyss fortio-deploy-b6757cbbb-7c9qg  -c fortio -- fortio load -c 50 -qps 0 -n 500 -loglevel Warning http://movies-service:8081/api/movies
```

Вывод будет типа такого

```bash
IP addresses distribution:
10.106.113.46:8081: 421
Code 200 : 79 (15.8 %)
Code 500 : 22 (4.4 %)
Code 503 : 399 (79.8 %)
```
Можно еще проверить статистику

```bash
kubectl exec -n cinemaabyss fortio-deploy-b6757cbbb-7c9qg -c istio-proxy -- pilot-agent request GET stats | grep movies-service | grep pending
```

И там смотрим 

```bash
cluster.outbound|8081||movies-service.cinemaabyss.svc.cluster.local;.upstream_rq_pending_total: 311 - столько раз срабатывал circuit breaker
You can see 21 for the upstream_rq_pending_overflow value which means 21 calls so far have been flagged for circuit breaking.
```

Приложите скриншот работы circuit breaker'а

Удаляем все
```bash
istioctl uninstall --purge
kubectl delete namespace istio-system
kubectl delete all --all -n cinemaabyss
kubectl delete namespace cinemaabyss
```
