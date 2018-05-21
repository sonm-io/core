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
          "id": "test_1",
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

After starting worker in dev mode on port `15010`, run:

```
$ make build/autocli && ./autocli worker start --input task.json --remote 8125721C2413d99a33E351e1F6Bb4e56b6b633FD@127.0.0.1:15010
```
