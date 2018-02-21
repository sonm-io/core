### Starting a task (dev mode)

Sample `task.json`:

```
{
  "container": {
    "image": "httpd:latest",
    "networks": [
      {
        "type": "l2tp",
        "options": {
          "lns_addr": "172.16.1.126",
          "subnet": "10.0.5.0/24",
          "ppp_username": "any",
          "ppp_password": "any"
        }
      }
    ]
  }
}
```

After starting miner in dev mode in port `15010`, run:

```
$ make build/autocli && ./autocli miner start --input task.json --remote 8125721C2413d99a33E351e1F6Bb4e56b6b633FD@127.0.0.1:15010
```
