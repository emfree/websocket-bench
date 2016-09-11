Informal concurrency benchmarks

Notes: In order to spawn large numbers of threads, you may need to increase
* The file descriptor limit (`ulimit -n`) and nproc limit (`ulimit -u`), e.g., by adding
  ```
  * soft nofile 131072
  * hard nofile 131072
  * soft nproc 131072
  * hard nproc 131072
  ```
  in `/etc/security/limits.conf`.

* `/proc/sys/kernel/threads-max` and `/proc/sys/kernel/pid_max`

* If using systemd, the `UserTasksMax` limit, the number of tasks that can be spawned from a login shell (in `/etc/systemd/logind.conf`).

You may also need to increase the ephemeral port range (`/proc/sys/net/ipv4/ip_local_port_range`) and enable tcp_tw_reuse (`/proc/sys/net/ipv4/tcp_tw_reuse`) in the client.
