# scopeweaver-cli 기여 안내

[English](CONTRIBUTING.md)

README.ko.md, docs/ko/cli.md, docs/ko/CLI-VERIFICATION.md를 읽으세요. AGPL-3.0·출처·로컬 격리 환경 사용 제한을 유지합니다. 기존 엔진 소스·테스트는 동작과 근거를 보존합니다.

dev에서 기능·수정 브랜치를 만들고 dev로 PR을 제출합니다. main 승격, 원격 저장소 생성, 공개에는 유지관리자 승인과 사람의 검토가 필요합니다. 관련 없는 작업을 보존하세요.

명령 계약은 internal/cli/operations.json입니다. 새 기능에는 이름 있는 명령, 매개변수·영향·응답 메타데이터, 소스 변경 감지 테스트가 필요합니다. 실제 동작으로 영향을 분류하고 통과를 위해 검증을 약화하지 마세요.

README·docs는 영어·한국어를 함께 유지합니다. 인터페이스 요약은 en·ko·zh-Hans를 지원하며 원본 증거·기술 진단을 보존합니다.

    go test ./internal/cli ./internal/portable ./packaging ./selfupdate -count=1
    go vet ./internal/cli ./internal/portable ./packaging ./selfupdate
    go run ./packaging -target windows-amd64

SCOPEWEAVER_CLI_E2E_BINARY를 패키지 실행 파일 경로로 지정한 뒤 실행합니다.

    go test ./internal/cli -run '^TestPortableCLIEndToEnd$' -count=1 -v

Linux·macOS에서는 해당 대상을 선택하세요. E2E는 다른 디렉터리에서 패키지 실행 파일과 임시 상태를 사용합니다. 핵심 DB 테스트는 ARTEX_TEST_DSN으로 지정한 새로운 scopeweaver_test_* 데이터베이스가 필요합니다. 사용자 DB를 쓰거나 실제 실행 검증을 모의 성공으로 바꾸지 마세요. 기존 실패·생략도 보고합니다.

런타임 변경은 인증, 동시 시작, 비정상 종료 복구, 정상 종료, 종료된 전체 폴더 이동, home별 스킬을 확인합니다. 버전 불일치·인증 정보 누락 시 데이터를 보존합니다.

패키지는 고정 런타임 체크섬·의존성 고지·파일 명세를 포함합니다. 공개 전 제3자 재배포 조건을 검토하세요. 자동 공개 워크플로는 없습니다. 기존 워크플로는 역사적 참고용 docs/upstream-workflows에만 보존합니다.
