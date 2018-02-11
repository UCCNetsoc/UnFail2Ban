echo "Binary name:"
read name
tsc static/ts/*.ts && go build -o $name *.go