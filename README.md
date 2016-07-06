# Diego Blame
## a cf cli plugin to find noisy neighbors on a given diego cell


### how to install
```bash
$ export myos=linux|osx|win
$ wget https://github.com/pivotalservices/diego-blame/releases/download/v1.0.0/diego-blame-${myos}
$ cf install-plugin diego-blame-${myos}
**Attention: Plugins are binaries written by potentially untrusted authors. Install and use plugins at your own risk.**

Do you want to install the plugin diego-blame? (y or n)> y

Installing plugin ./diego-blame...
OK
Plugin diego-blame v1.0.0 successfully installed.

$ cf plugins
Listing Installed Plugins...
OK

Plugin Name                     Version   Command Name   Command Help
diego-blame                     1.0.0     diego-blame    Run a scan to find all apps on a given diego cell in order to identify utilization spike causes

```



### what it looks like in action
```bash
$ cf diego-blame 192.xxx.x.201
+----------------------------------------------+---------+---------------------+----------------+-----------------+----------------+-----------------+------------------------+--------------+------------------------------------------------------------------------+
|                   APPNAME                    |  STATE  |      HOST:PORT      |   DISK-USAGE   |   DISK-QUOTA    |   MEM-USAGE    |    MEM-QUOTA    |       CPU-USAGE        |    UPTIME    |                                  URIS                                  |
+----------------------------------------------+---------+---------------------+----------------+-----------------+----------------+-----------------+------------------------+--------------+------------------------------------------------------------------------+
| xxxxx-sssssss                                | RUNNING | 192.xxx.x.201:60131 | 1.27451136e+08 | 1.073741824e+09 | 3.99814656e+08 | 5.36870912e+08  |  0.0006871335636523104 | 1.210252e+06 | [xxxxx-pcfdemo.cfapps.xxx.pivotal.io]                                  |
| aaaa                                         | RUNNING | 192.xxx.x.201:60140 | 6.221824e+06   | 1.073741824e+09 | 1.1276288e+07  | 3.3554432e+07   | 0.00036162815279443836 | 1.209784e+06 | [xxxx.cfapps.xxx.pivotal.io]                                           |
| oooooooooooo                                 | RUNNING | 192.xxx.x.201:60114 |         475136 | 1.073741824e+09 | 1.140736e+07   | 1.073741824e+09 |  0.0003231596649347209 | 1.210601e+06 | [xxxxxx-hello.cfapps.xxx.pivotal.io]                                   |
+----------------------------------------------+---------+---------------------+----------------+-----------------+----------------+-----------------+------------------------+--------------+------------------------------------------------------------------------+
```
