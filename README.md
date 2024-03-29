# 진원봇 : 바라미실 문 개폐 알림 봇
진원봇(jinwonbot)은 바라미실 문 개폐 여부를 원격지에서 확인할 수 있는 채팅 봇입니다. 

## 개발 동기
바라미실의 출입문이 열려있는지의 여부는 준회원의 동아리실 출입을 도와주거나 재실 인원이 있는지 확인하기 위해 지속적인 수요가 존재해왔습니다.
이 때문에 이전부터 *바라미캠(2016, 송유진)* 과 *바라미는 열렸는가?(2017, 김상윤 외 2명)* 와 같이 바라미실 문 개폐 여부를 확인하기 위한 다양한 형태의 작품이 출품되어왔습니다.
진원봇은 이러한 시도를 집대성해 쉽게 접근할 수 있고 정확하고 다양한 정보를 제공하는 것에 초첨을 맞춘 작품입니다.

## 설계
개발 동기에도 언급하였듯이, 본 작품의 설계에서 아래와 같은 이전 작품들의 시행착오를 많이 참고했습니다. 
- **바라미캠**의 경우, 웹 프론트엔드의 요청이 발생하면 라즈베리파이 백앤드에서 웹캠으로 이미지를 촬영하고 이를 정적 파일로 바라미 서버에 업로드하면 이를 웹 서버가 서빙하는 구조로 보입니다.
이런 구조의 경우 제공된 사진을 사람이 직접 해석하여 개폐 여부를 판단해야 했습니다. (물론 이 해석 과정을 자동화하는 것도 불가능하진 않겠지만 절대 쉬운 과정은 아닐 것 같습니다.)
- **바라미는 열렸는가?** 의 경우, SECOM 보안 단말기의 문 잠금 여부 표시등에 CDS를 부착해 점등 여부로 개폐 여부를 결정하고 라즈베리파이에서 정적 파일을 생성해 웹 서버가 서빙하는 구조로,
CDS가 표시등 이외 주변 불빛을 감지하지 못하게 방지하는 차폐의 문제로 가끔 올바르게 감지하지 못하는 문제가 있었습니다.

위에 언급한 것과, 다른 부가적인 시행착오를 바탕으로 도출할 수 있었던 설계 중점과 그 결과는 아래와 같았습니다.
- 채팅 봇의 생명인 접근성을 확보하기 위해 다양한 형태의 프론트엔드 *(예: 웹, 카카오톡 플친, 디스코드)* 를 구현할 수 있는 구조 : 백앤드와 프론트앤드가 격리된 구조, 범용성있는 HTTP REST API
- 정확한 개폐 상태를 확인할 수 있는 센서 : 간섭하기 그나마 어려운 자기장 감지
- 현재 상태만 확인하는 것에서 국한되지 않고 통계와 같은 다양한 기록을 제공하기 위해 과거의 기록을 유지 : 백앤드에 상태를 저장하는 DB 포함
- 많은 요청 수 처리 성능 : 스레드 안전한 요청 핸들러 구현하고 고루틴(golang에서 제공하는 일종의 경량 스레드)을 적극 활용 
- 서버 운영 안정성과 편의성 충족 : 백앤드 언어로 거의 모든 OS, 아키텍처 크로스 컴파일 지원하는 golang을 채택하고 컨테이너 가상화

## 구조
본 작품은 크게 문 개폐 여부를 센싱하는 센서와 센싱 정보를 처리하여 저장하고 API를 통해 정보를 제공하는 백앤드, 받아온 정보를 가공하여 보기 쉽게 표시하는 프론트앤드 세 부분으로 나뉩니다.
![project architecture diagram](demo/structure.png) 

- **센서**: 전자석으로 문을 잠그는 SECOM 단말기 잠금 장치에 9축 IMU센서인 MPU9050을 부착하고 파이썬 데몬이 일정 시간마다 센서의 자기장 평균값을 읽어 상태가 바뀌었다고 판단되면 정해진 깃헙 저장소에 커밋을 수행합니다. 이때 커밋 메세지에 감지 시간과 문 상태가 담기며 이 데이터로 생성된 정적 웹 페이지가 커밋됩니다.
- **백엔드**: 일정시간마다 백앤드 데몬은 저장소의 커밋을 확인해 로컬 DB (JSON store)와 원격 저장소의 커밋 목록을 동기화합니다. echo 웹 서버와 디스코드 메세지 처리기는 요청이 들어올때마다 로컬 DB에 질의를 수행해 응답합니다.
- **프론트엔드**: 현재 상태로썬 두 개의 일반 사용자용 프론트앤드와 프론트 앤드를 제작하기 위한 REST API가 구현되어 있습니다.

## 시연 사진
디스코드 프론트엔드 (바라미 디스코드방이나 진원쿤#5014에게 직접 메세지를 보내 사용해보실 수 있습니다.)
![Discord bot demo](demo/discord1.png)
![Discord bot demo](demo/discord2.png)

웹 프론트엔드 (<a href="https://ibarami.github.io" target="_blank"> https://ibarami.github.io</a> 에서 직접 사용해보실 수 있습니다.)
![WEB frontend demo](demo/web1.png)

REST API 예시 (<a href="https://api.chinchister.com/jinwonbot/" target="_blank">https://api.chinchister.com/jinwonbot/</a>에서 직접 사용해보실 수 있습니다. <a href="https://github.com/Dictor/jinwonbot/blob/master/docs/api.md" target="_blank">API 명세서</a> )
   
![REST API demo](demo/rest.png)

## FAQ
- **왜 백앤드에서 센서 감지까지 한 번에 수행하지 않고 센서 감지부와 백앤드가 분리된 구조인가요?** : 원래 진원 선배가 제작한 센서와 웹 프론트엔드만 존재하던 독립적인 작품이였습니다. 거기에 제가 원래 작품의 구조를 유지하면서 추가적으로 여러 가지 기능을 구현하다보니 지금과 같은 구조를 가지게 되었습니다. 덕분에 복잡성은 조금 증가했지만 개폐 기록에 대한 데이터베이스와 프론트앤드를 제공하는 서버가 이중 구조를 띄어 내고장성이 향상되었습니다.
- **이 DB를 선택한 이유가 있나요?** : 사용한 DB의 종류가 총 3번이 바뀌었습니다. 컨테이너로 가상화하면서 간단한 서버에 또다른 DB 서버를 붙이는건 오버헤드가 큰거 같아서 임베디드 DB만 고려했습니다. 가장 처음엔 sqlite3 + godb (sql query builder) 구성이였는데, sqlite3 코어가 순수 golang으로 구현되지 않아 cgo의 개입이 필연적이였습니다. 이는 크로스컴파일에 큰 악재로 작용했고, 순수 golang으로만 구현된 DB를 찾다가 발견한것이 kv-store인 badgerDB입니다. 만족스러운 성능을 보여준 badger이지만 그 복잡성으로 인해 정교하게 맞춰진 설정이 없으면 DB가 깨지는 현상을 겪으면서 차라리 최소기능에 간단한 store를 구현하자 싶어서 golang 표준 라이브러리인 json (언)마샬러를 이용해 필요한 간단한 기능만 구현해 쓴 것이 지금에 이르게 되었습니다.

## 개선할 점
- 접근성과 편의성을 위한 더 다양한 종류의 프론트앤드를 구현해야 합니다. 기존 프론트앤드의 개선도요.

## 소스코드

- <a href="https://github.com/Dictor/jinwonbot" target="_blank">백앤드, 디스코드 프론트엔드 및 REST API 소스</a>
- <a href="https://github.com/ibarami/ibarami.github.io" target="_blank">웹 프론트엔드 소스</a>
- <a href="https://github.com/ibarami/IsBaramiOpen" target="_blank">센서 소스</a>

## 감사한 분
작품과 디스코드 프론트앤드의 이름과 프로필사진, 하드웨어를 제공해주신 24기 주진원 선배님께 다시 한번 감사드립니다! 그리고 디스코드 봇 사용해주신 바라미 회원분들도 완전 감사!
