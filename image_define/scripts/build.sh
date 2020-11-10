docker build -t sugar/appserver:latest -f ./scripts/predictor.dockerfile $PWD 
docker build -t sugar/trainer:latest -f ./scripts/train.dockerfile $PWD 