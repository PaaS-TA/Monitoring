# PaaS-TA-Monitoring

**파스-타 모니터링 구성은 다음과 같다.**

<table>
<tr>
  <td>패키지명</td>
  <td>설명</td>
</tr>
<tr>
  <td>paasta-controller-2.0</td>
  <td>paasta 2.0 controller 서비스 릴리즈</td>
</tr>
<tr>
  <td>paasta-container-2.0</td>
  <td>paasta 2.0 container 서비스 릴리즈</td>
</tr>
<tr>
  <td>paasta-garden-runc-2.0</td>
  <td>paasta 2.0 garden runc 서비스 릴리즈</td>
</tr>
<tr>
  <td>paasta-influxdb-grafana-2.0</td>
  <td>paasta 2.0 매트릭스 데이터베이스 시스템(influxdb)와 화면 UI (grafana)로 구성된 릴리즈 </td>
</tr>
<tr>
  <td>paasta-logsearch-2.0</td>
  <td>paasta 2.0 로그 정보 관리 시스템(logsearch) 릴리즈.</td>
</tr>
<tr>
  <td>paasta-metrics-collector-2.0</td>
  <td>paasta 서비스관련 매트릭스를 수집하여 데이터베이스(influxdb)에 저장하는 서비스 릴리즈</td>
</tr>
<tr>
  <td>paasta-monitoring-agent-2.0</td>
  <td>Bosh 및 외부 시스템 관리를 위한 매트릭스 수집 Agent 프로젝트 소스</td>
</tr>
<tr>
  <td>paasta-alarm-service-2.0</td>
  <td>paasta 2.0 관련 서비스들의 임계치 모니터링 및 알람 서비스 릴리즈</td>
</tr>
</table>

- 파스타 모니터링 시스템을 사용하기 위해서는 기본적으로 Bosh 서비스와 PaaS-TA Controller 및 Container 2.0 서비스가 설치되어 있어야 한다.
- 3가지 타입의 Bosh - bosh-lite, micro-bosh, full-bosh- 중에 bosh-lite 타입은 모니터링 시스템에서 지원하지 않는다.
- Bosh 및 PaaS-TA Controller 및 Container 2.0 서비스 설치는 [플랫폼 자동화 설치](https://github.com/PaaS-TA/Guide-2.0-Linguine-/blob/master/Install-Guide/Platform%20Install%20System/PaaS-TA_%ED%94%8C%EB%9E%AB%ED%8F%BC_%EC%84%A4%EC%B9%98_%EC%9E%90%EB%8F%99%ED%99%94_%EC%84%A4%EC%B9%98_%EA%B0%80%EC%9D%B4%EB%93%9C.md)를 참조한다.
- Bosh 및 PaaS-TA Controller 및 Container 2.0 서비스 설치 완료 후 아래 서비스를 설치한다.
 1. paasta-influxdb-grafana-2.0 서비스 설치
 2. paasta-logsearch-2.0 서비스 설치
 3. paasta-metrics-collector-2.0 서비스 설치
 4. paasta-monitoring-agent-2.0 서비스 설치 <br>
    [Bosh Monitoring Agent 설치 참조](https://github.com/PaaS-TA/Guide-2.0-Linguine-/blob/master/Install-Guide/BOSH/Bosh%20Monitoring%20Agent%20%EC%84%A4%EC%B9%98%20%EA%B0%80%EC%9D%B4%EB%93%9C_v1.0.md)
 5. paasta-alarm-service-2.0 서비스 설치
