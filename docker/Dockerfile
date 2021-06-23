FROM ubuntu:xenial
LABEL version="1.0.1"

# set apt mirrors
RUN sed -i s@/archive.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list \
    && apt-get clean \
    && apt-get update \
    && mkdir /home/openjudge

# add gcc g++ environment
RUN apt-get install -y build-essential
ENV PATH /usr/lib:$PATH

# add java environment
ADD jdk-8u291-linux-x64.tar.gz /usr/local/java
RUN ln -s /usr/local/java/jdk1.8.0_291 /usr/local/java/jdk
ENV JAVA_HOME /usr/local/java/jdk 
ENV JRE_HOME ${JAVA_HOME}/jre 
ENV CLASSPATH .:${JAVA_HOME}/lib:${JRE_HOME}/lib 
ENV PATH ${JAVA_HOME}/bin:$PATH

# add go environment
ADD go1.16.4.linux-amd64.tar.gz /usr/local
ENV PATH /usr/local/go/bin:$PATH

ENV TZ=Asia/Shanghai

# add Judger
COPY gojudger /home/openjudge/judger

# run judger
CMD /home/openjudge/judger

