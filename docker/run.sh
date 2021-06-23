#! /bin/bash

USERNAME="openjudge"
HOME_PATH="/home/$USERNAME"
if ! id -u $USERNAME >/dev/null 2>&1; then
	echo "create user"
	sudo useradd -d "/home/$USERNAME" -m $USERNAME
	echo "create user success， please input user password"
fi
echo 

USERID=`id -u $USERNAME`
echo "$USERNAME's user id = $USERID"

if [ ! -d $HOME_PATH ]; then
	sudo mkdir $HOME_PATH
fi

sudo cp openjudge.conf "$HOME_PATH/conf/openjudge.conf"

KERNEL_VERSION=`uname -r`
echo "kernel version: $KERNEL_VERSION"

docker run --privileged=true  -it \
-u $USERID \
--ulimit nproc=100 \
-v $HOME_PATH/data:$HOME_PATH/data \
-v $HOME_PATH/work:$HOME_PATH/work \
-v $HOME_PATH/conf:$HOME_PATH/conf \
-v $HOME_PATH/go-build:/.cache/go-build \
-v /boot/config-$KERNEL_VERSION:/boot/config-$KERNEL_VERSION \
15692327692/oj:1.0.4

