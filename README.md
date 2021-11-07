# NNS back
![image](https://user-images.githubusercontent.com/44018094/140633883-68688082-2632-4bcb-acb5-7a011c40f98b.png)

[Neural Network Studio](https://nnstudio.io/) 의 API 서버

</br>

### Document
[Postman Document 바로가기](https://documenter.getpostman.com/view/13966682/Tzm5GwnJ)

</br>

### Back-end Architecture

![image](https://user-images.githubusercontent.com/44018094/140633790-e43cc48c-a914-4bec-af75-8337faee2612.png)

</br>

### Directory structure
```
├─cloud
├─dataset
│  └─testdata
│      └─zip
│          ├─train
│          └─validate
├─datasetConfig
├─externalAPI
├─log
├─model
├─repository
├─service
├─sql
├─train
├─util
└─ws
    └─message
```
- cloud : API 서버에서 사용하는 cloud 서비스 (AWS) API를 사용하기 편리하도록 Wrapping한 패키지
- dataset : 데이터셋 스토어 및 데이터셋 라이브러리 구현 패키지
- datasetConfig : 프로젝트 내의 데이터셋 설정 구현 패키지
- externalAPI : API 서버에서 사용하는 외부 API를 Wrapping한 패키지
- log : Go언어의 유명 log 라이브러리인 [uber-go/zap](https://github.com/uber-go/zap) 를 Wrapping한 패키지
- model : 프로젝트, 멤버, 이미지 등등 서비스에서 사용하는 도메인의 모델
- repository : 프로젝트, 멤버, 이미지 등등 서비스에서 사용하는 도메인의 인터페이스
- service : API 서버 서비스 구현 패키지
- sql : 모델 ddl 저장용 패키지
- train : 딥러닝 모델 학습, 학습 이력과 관련된 기능을 구현한 패키지
- util : 각종 유틸리티 함수 패키지
- ws : 웹소켓 서버를 구현한 패키지
  + message : 웹소켓 통신에 사용하는 메세지를 정의

</br>

### Build & Deploy
nns_back package의 root directory에서 `go build`. 단, `environment` 변수 설정 필요.   
서버 부팅시 자동으로 실행될 수 있도록 linux systemd service로 띄웠다. 해당 스크립트는 다음과 같다.
```
[Unit]
Description=Neural Network Studio API server

[Service]
Type=simple
Environment=DBUSER=***
Environment=DBPW=***
Environment=DBIP=***
Environment=DBPORT=***
Environment=AWS_SECRET_ACCESS_KEY=***
Environment=AWS_ACCESS_KEY_ID=***
Environment=IMAGE_BUCKET_NAME=***
Environment=DATASET_BUCKET_NAME=***
Environment=TRAINED_MODEL_BUCKET_NAME=***
WorkingDirectory=***
StandardOutput=***
StandardError=***
ExecStart=***
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

</br>

### CI/CD
Jenkins를 사용하려 했으나, 소규모 프로젝트 운영에 Jenkins용 서버를 하나 더 관리하는 것은 비용적 부담이 있다. 따라서 일정 사용량 이내에서는 무료로 사용 가능하고 서버를 직접 관리할 필요가 없는 **Github Action**을 사용하여 `master branch`에 push할 경우 AWS EC2 서버에 자동으로 배포되도록 설정했다.
