MAELSTROM_BIN=./maelstrom/maelstrom
GO_BIN=~/go/bin

01:
	cd 01-echo && go install .
	$(MAELSTROM_BIN) test -w echo --bin $(GO_BIN)/01-echo --time-limit 5

02:
	cd 02-unique-id && go install .
	$(MAELSTROM_BIN) test -w unique-ids --bin $(GO_BIN)/02-unique-id --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition

03a:
	cd 03a-single-broadcast && go install .
	$(MAELSTROM_BIN) test -w broadcast --bin $(GO_BIN)/03a-single-broadcast --node-count 1 --time-limit 20 --rate 10

03b:
	cd 03b-multi-broadcast && go install .
	$(MAELSTROM_BIN) test -w broadcast --bin $(GO_BIN)/03b-multi-broadcast --node-count 5 --time-limit 20 --rate 10

03c:
	cd 03c-fault-tolerant-broadcast && go install .
	$(MAELSTROM_BIN) test -w broadcast --bin $(GO_BIN)/03c-fault-tolerant-broadcast --node-count 5 --time-limit 20 --rate 10 --nemesis partition

03d:
	cd 03d-efficient-broadcast && go install .
	$(MAELSTROM_BIN) test -w broadcast --bin $(GO_BIN)/03d-efficient-broadcast --node-count 25 --time-limit 20 --rate 100 --latency 100

03e:
	cd 03e-efficient-broadcast && go install .
	$(MAELSTROM_BIN) test -w broadcast --bin $(GO_BIN)/03e-efficient-broadcast --node-count 25 --time-limit 20 --rate 100 --latency 100 --nemesis partition
