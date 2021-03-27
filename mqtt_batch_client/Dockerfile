#基于的镜像不是golang(733M),而是alpine(4.14M)
FROM alpine:latest

WORKDIR $GOPATH/src/bench

#此目录存放的是beego项目的配置
#ADD ./properties.json $GOPATH/src/bench/properties.json

# CA 证书
ADD ./root-cacert.pem $GOPATH/src/bench/root-cacert.pem
ADD ./client.key $GOPATH/src/bench/client.key
ADD ./client-cert.pem $GOPATH/src/bench/client-cert.pem

#编译后的可执行文件
ADD cp7_bench_sub_go $GOPATH/src/bench/cp7_bench_sub

#RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

EXPOSE 19876

ENTRYPOINT ./cp7_bench_sub ${args}