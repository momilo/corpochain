#!/usr/bin/env bash

cd ..
docker-compose up > logfile &
sleep 5
genesisAddress=$(grep 'wallets.go:55: generated address:' logfile | awk '{print $6}')
echo "genesis address: $genesisAddress"

curl http://localhost:8080/getBalance/$genesisAddress

sleep 1

newAddress=$(curl http://localhost:8080/createWallet | sed -e 's/address:"//' | sed -e 's/"//' | sed -e 's/ //')
echo "new address: $newAddress"

sleep 2

curl http://localhost:8080/getBalance/$newAddress

sleep 1


post_data()
{
cat <<EOF
 {"fromAddress":{"address":"$genesisAddress"}, "toAddress":{"address":"$newAddress"}, "amount":{"amount":$1}}
EOF
}

echo "Balances"
curl http://localhost:8080/getBalance/$genesisAddress
curl http://localhost:8080/getBalance/$newAddress

#-------------------------------------------------

echo "Sending 100 from $genesisAddress to $newAddress"

curl http://localhost:8080/send -X POST -d "$(post_data 100)"

sleep 1

curl http://localhost:8080/getBalance/$genesisAddress
curl http://localhost:8080/getBalance/$newAddress

#-------------------------------------------------

echo "Sending 100 from $genesisAddress to $newAddress"

curl http://localhost:8080/send -X POST -d "$(post_data 100)"

sleep 1

curl http://localhost:8080/getBalance/$genesisAddress
curl http://localhost:8080/getBalance/$newAddress

#-------------------------------------------------

echo "Sending 300 from $genesisAddress to $newAddress"

curl http://localhost:8080/send -X POST -d "$(post_data 300)"

sleep 1

curl http://localhost:8080/getBalance/$genesisAddress
curl http://localhost:8080/getBalance/$newAddress

#-------------------------------------------------

echo "Sending 1300 from $genesisAddress to $newAddress"

curl http://localhost:8080/send -X POST -d "$(post_data 1300)"

sleep 1

curl http://localhost:8080/getBalance/$genesisAddress
curl http://localhost:8080/getBalance/$newAddress