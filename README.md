## DNS proxy with domain based forwarding

К примеру вы подключены к корпоративной сети и в этом случае все DNS запросы уходят на корпоративный DNS сервер.
Чтобы этого избежать и отправлять на корпортивный DNS сервер только нужные ему запросы как раз и пригодится этот DNS прокси.
На порту 9970 доступны метрики в prometheus формате.

### Пример конфигурации
```
nameservers:
  - 192.168.100.61:53
  - 8.8.8.8:53
corpnameservers:
  - 10.10.10.10:53
corpdomain: example.com
excludecorpdomain: vpn.example.com
blocklist:
  - https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts
blockAddress4: 0.0.0.0
blockAddress6: 0:0:0:0:0:0:0:0
configUpdate: true
updateInterval: 12h
```

`nameservers` - куда отправлять все DNS запросы

`corpnameservers` - адрес корпоративного DNS сервера (только один сервер поддерживается)

`corpdomain` - на основе этого домена будет происходить фильтрация запросов которые нужно отправить на корпоративный DNS сервер

`excludecorpdomain` - это адрес VPN сервера к которому подключается VPN. Его адрес мы должны получить через публичный DNS, иначе VPN не заработает.

`blocklist` - адрес откуда можно скачать список доменов для блокировки формат файла аналогичный файлу /etc/hosts.

`blockAddress4`, `blockAddress6` - адреса, на которые будут заменены домены из blocklist.

`configUpdate` - динамическое обновление конфигурации, если вдруг она поменяется

`updateInterval: 12h` - интервал с которым нужно обновлять blocklist.

### Как запускать 
```
docker build -t dns:1.0 . 
docker run -d -p 53:53 -p 53:53/udp -p 9970:9970 \
      --name mydns \
      --rm \
      -v "./config.yaml:/config.yaml" \
     dns:1.0
```
После запуска контейнера нужно заменить локальный DNS сервер на 127.0.0.1 если контейнер запущен локально.
Если контейнер запущен где-то еще, то корпоративный DNS должен быть доступен с того места где работает контейнер.

### Пример работы
```
docker logs -f mydns
2024/06/25 13:02:16 qName: github.com. routed to public DNS
2024/06/25 13:03:54 qName: vpn.example.com. routed to public DNS due to corporate exclude domain
2024/06/25 13:04:28 qName: git.example.com. routed to corp DNS
```