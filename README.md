Informal concurrency benchmarks

Notes: In order to spawn large numbers of threads, you may need to increase
* The file descriptor limit (`ulimit -n`) and nproc limit (`ulimit -u`), e.g., by adding
  ```
  * soft nofile 131072
  * hard nofile 131072
  * soft nproc 131072
  * hard nproc 131072
  ```
  in /etc/security/limits.conf.

* `/proc/sys/kernel/threads-max` (`sysctl kernel.threads-max`).

* If using systemd, the `UserTasksMax` limit, the number of tasks that can be spawned from a login shell (in `/etc/systemd/logind.conf`).
