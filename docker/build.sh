CURRENT_PATH=`pwd`
PROJECT_PATH=/home/zhaoyu/桌面/OpenJudge/GOJudger
echo $CURRENT_PATH
go env -w GOROOT=/usr/local/go
go env -w GOPATH=$PROJECT_PATH #gosetup
/usr/local/go/bin/go build -o $CURRENT_PATH/gojudger $PROJECT_PATH/src/main/main.go

