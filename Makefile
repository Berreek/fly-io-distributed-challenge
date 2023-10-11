.PHONY: echo
echo:
	cd ./pkg/maelstrom-echo && go install .
	./maelstrom/maelstrom test -w echo --bin ~/go/bin/maelstrom-echo --node-count 1 --time-limit 10

.PHONY: unique-id
unique-id:
	cd ./pkg/maelstrom-unique-id && go install .
	./maelstrom/maelstrom test -w unique-ids --bin ~/go/bin/maelstrom-unique-id --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition

.PHONY: broadcast-single-node
broadcast-single-node:
	cd ./pkg/maelstrom-broadcast && go install .
	./maelstrom/maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 1 --time-limit 20 --rate 10

.PHONY: broadcast-multi-node
broadcast-multi-node:
	cd ./pkg/maelstrom-broadcast && go install .
	./maelstrom/maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 5 --time-limit 20 --rate 10

.PHONY: broadcast-partition
broadcast-partition:
	cd ./pkg/maelstrom-broadcast && go install .
	./maelstrom/maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 5 --time-limit 20 --rate 10 --nemesis partition

.PHONY: broadcast-efficient
broadcast-efficient:
	cd ./pkg/maelstrom-broadcast && go install .
	./maelstrom/maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 25 --time-limit 20 --rate 100 --latency 100

.PHONY: broadcast-efficient-with-partition
broadcast-efficient-with-partition:
	cd ./pkg/maelstrom-broadcast && go install .
	./maelstrom/maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 25 --time-limit 20 --rate 100 --latency 100 --nemesis partition
