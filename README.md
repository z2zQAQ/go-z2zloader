# go-z2zloader

an Anti- Virus demo using go

# 介绍

学习了一段时间go和免杀 做了个小demo，本来计划是cobra库做成工具使用的，但是后来觉得直接编译成exe比较方便，就上传了这个旧的版本。

静态加密使用的是aes+base64

动态直接调用syscal底层函数，挺老的办法没想到还能用，还尝试了APC注入，但是有时候上不了线

加入了junkcode和anti-vm模块，还在测试中



# 使用

cs stageless生成raw shellcode

![image-20241227124903604](README/image-20241227124903604.png)

```
首先在main函数中调用encode模块 加密原生shellcode
然后选择使用本地（Original）还是远程（remote）模块，远程模块也是在开启web服务的目录下放置一个加密后的shellcode

1.不注释第11行，输入原生shellcode加密
go run main.go 
2.选择加密模式，输入加密后的bin文件
go build -ldflags="-s -w" -o 输出exe文件名 main.go
```

![image-20241227124815484](README/image-20241227124815484.png)

# 免杀效果

加密后的bin文件

![image-20241227115947651](README/image-20241227115947651.png)

生成的exe过火绒 df 360

![image-20241226151645350](README/image-20241226151645350.png)



![image-20241226161253029](README/image-20241226161253029.png)



上线

360

![image-20241227112249120](README/image-20241227112249120.png)

![image-20241227112324291](README/image-20241227112324291.png)

核晶

![image-20241227112649282](README/image-20241227112649282.png)

火绒

![image-20241227114234152](README/image-20241227114234152.png)



exe沙箱结果

![image-20241226162856494](README/image-20241226162856494.png)



# 后续

APC模块昨天测试的时候会被360杀，今天上线不了了，还在排查原因

添加免杀模块，尝试过沙箱

免黑框



# 学习资料

[Junk-Go-Generator/Junk Code Generation.go at master · SaturnsVoid/Junk-Go-Generator](https://github.com/SaturnsVoid/Junk-Go-Generator/blob/master/Junk Code Generation.go)

[go实现免杀(实用思路篇) | CN-SEC 中文网](https://cn-sec.com/archives/2839255.html)

[（●´3｀●）好啦好啦](https://shut-td.github.io/CS远控免杀思路与实现/)

[免杀技术 - go shellcode 加载 bypassAV | Hyyrent blog](https://pizz33.github.io/posts/4ac17cb886a9/)

https://github.com/HZzz2/go-shellcode-loader
