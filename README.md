<img src="images/dlam.png" width="18%" height="18%"/>

Dlam可从**hunter**、**quake**、**fofa**等**网络空间测绘平台收集探测互联网IP**，检测其中可使用的IP，并通过本工具配置文件中的端口映射关系，把本地端口映射到互联网IP指定的端口，以便通过互联网IP访问本地端口服务。本工具基于[frp](https://github.com/fatedier/frp)修改。
> **开发不易,如果觉得该项目对您有帮助, 麻烦点个Star吧**

### 0x01 说明

1. 本工具主要解决通过互联网IP访问本地端口服务的问题，即：把本地端口映射到某些互联网IP的端口上
2. 适合无固定互联网IP时使用
3. 适合防溯源时使用
4. 不要把本地有漏洞的服务映射到互联网

### 0x02 使用

1、修改myPortList的值为自己想映射到互联网IP上使用的端口，此处的端口为互联网IP的端口，需要避免可能已被使用的端口。检测互联网IP是否可用，后续使用过程中访问本地服务时，通过互联网IP:此处的PORT来访问

![image-20240907220246285](images/myPortList.png)



2、修改测绘平台key信息，使用哪个，就修改哪个的switch='open'

![image-20240907215918404](images/netspace.png)



3、./dlam check命令执行程序，获取可用的互联网IP

![image-20240907221225651](images/check.png)



4、执行check命令后，此时，可以看到config.toml中的whoseInternetAddress已存入了可用的互联网IP，此时可以关闭各空间测绘的开关，后续再执行check命令时，会自动取此处的互联网IP来检测，当此处的IP都不可用时，再打开测绘平台的开关重新获取即可。

![image-20240907221546747](images/internetAddress.png)



5、在config.toml中配置dnats，把本地端口和对应的互联网IP:PORT对应起来（有多种对应形式）,举例如下

![image-20240907222614903](images/dnats.png)

6、配置好之后，直接./dlam无命令启动程序，就可以通过互联网IP:PORT来访问本地服务了

![image-20240907222957089](images/start.png)

### 0x03 效果举例

1、cs反连

本地端口远程端口保持一致

<img src="images/csport.png" alt="image-20240907223629545" style="zoom:50%;" width="35%"  height="35%"/>

启动监听,此处IP要设置公网IP

<img src="images/cslistener.png" alt="image-20240907223837918" style="zoom:50%;" width="35%" height="35%"/>

监听如下：

![image-20240907224019705](images/cslistener2.png)

执行payload，反连成功

![image-20240907224731516](images/cssuccess.png)

如以往一样操作：

<img src="images/cstest.png" alt="image-20240907224929592" style="zoom:33%;" width="50%" height="50%" />

2、nc反连

![image-20240907230208360](images/nc.png)

3、其他同理......
