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

.PHONY: counter
counter:
	cd ./pkg/maelstrom-counter && go install .
	./maelstrom/maelstrom test -w g-counter --bin ~/go/bin/maelstrom-counter --node-count 3 --rate 100 --time-limit 20 --nemesis partition

.PHONY: kafka
kafka:
	cd ./pkg/maelstrom-kafka-style-log && go install .
	./maelstrom/maelstrom test -w kafka --bin ~/go/bin/maelstrom-kafka-style-log --node-count 2 --concurrency 2n --time-limit 20 --rate 1000

.PHONY: totally-available
totally-available:
	cd ./pkg/maelstrom-totally-available && go install .
	./maelstrom/maelstrom test -w txn-rw-register --bin ~/go/bin/maelstrom-totally-available --node-count 2 --concurrency 2n --time-limit 20 --rate 1000 --consistency-models read-committed --availability total --nemesis partition
