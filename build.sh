echo "Binary name:"
read name
tsc && go build -o $name *.go