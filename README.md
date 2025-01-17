# 교배식 최단 경로 찾기

본 프로그램은 Go + Fyne 으로 작성된 GUI 애플리케이션이며,
"교배식.txt" 파일을 분석해, 특정 '부모 팰'에서 '자식 팰'로 이어지는 최단 교배 경로를 찾아줍니다.

--------------------------------------------------------------------------------
1. 기능 요약
--------------------------------------------------------------------------------
- 교배식 파일 열기:
  "부모+부모=자식" 형태의 교배식들을 읽어, 중복/충돌 교배식을 검사하고
  내부 데이터로 구성합니다.

- 교배 경로 검색:
  부모 팰, 자식 팰을 입력하면 BFS로 최소 교배 횟수 경로들을 찾아 보여줍니다.

- 교배식 추가:
  새 교배식을 입력해 파일에 한 줄 추가하고,
  즉시 데이터 재로딩 후 검색에 반영합니다.

- Windows에서 콘솔 없이 GUI 실행:
  빌드시 -ldflags "-H=windowsgui" 옵션 사용.
  rsrc(또는 go-winres) 등을 통해 .syso 파일을 만들면 exe 아이콘 변경 가능.

--------------------------------------------------------------------------------
2. 사전 준비
--------------------------------------------------------------------------------
- Go 환경 (1.18 이상 권장) 설치
- Fyne 라이브러리 설치:
  go get fyne.io/fyne/v2
- 교배식.txt 파일 (UTF-8 인코딩) 준비

--------------------------------------------------------------------------------
3. 교배식 예시
--------------------------------------------------------------------------------
꼬꼬닭+꼬꼬닭=꼬꼬닭
까부냥+까부냥=까부냥
다크울프+마그피스=드라이지스
꼬꼬닭+드라이지스=불이리
...

형식:
부모1+부모2=자식
구분자는 +, =, 공백.
중복/충돌 발생 시 에러 처리.

--------------------------------------------------------------------------------
4. 빌드 및 실행
--------------------------------------------------------------------------------
1) go mod init my-breeding
   go mod tidy

2) (선택) 아이콘 리소스 생성 (예: rsrc)
   go install github.com/akavel/rsrc@latest
   rsrc -ico myicon.ico -o rsrc.syso

3) 빌드 (Windows GUI, 콘솔 숨김)
   go build -ldflags "-H=windowsgui" -o breeding.exe

4) 실행
   breeding.exe (탐색기에서 더블클릭)

--------------------------------------------------------------------------------
5. 사용 방법
--------------------------------------------------------------------------------
1) 프로그램 실행 → "교배식 파일 열기" → 교배식.txt 선택
2) 부모 팰, 자식 팰 입력 후 "검색" 클릭
3) 검색 결과가 텍스트 영역에 표시
4) 필요 시 "교배식 추가" 버튼 → 새 (부모+부모=자식) 입력 → 확인
   → 파일에 추가 후 재로딩되어 바로 검색 가능

--------------------------------------------------------------------------------
6. 기타 주의사항
--------------------------------------------------------------------------------
- Windows 아이콘 캐시 문제로, 아이콘 변경이 즉시 반영 안 될 수 있음
  → 다른 파일명으로 빌드하거나 탐색기 새로고침 필요
- 대규모 데이터나 복잡한 교배식은 속도 이슈가 있을 수 있음
- 중복/충돌 교배식 발견 시 에러로 검색 중단

--------------------------------------------------------------------------------
7. 라이선스
--------------------------------------------------------------------------------
- 예제 코드는 자유롭게 수정, 재배포 가능
- Go, Fyne 등은 각자 라이선스에 따라 사용
