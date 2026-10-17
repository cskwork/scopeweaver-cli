# 독립 실행형 CLI 검증

[English](../CLI-VERIFICATION.md)

2026-10-17 Windows amd64에서 기록했습니다. 엔진 기준은 ScopeWeaver 커밋2df0b307dcb817a459abebd3a4c240a2019598db입니다. 임시 로컬 상태와 PostgreSQL17.11을 사용했고 실제 대상·유료 모델·외부 알림을 호출하지 않았습니다.

## 통과

- 명령297개·매개변수857개. 모든 API 등록 경로, 생성되는 별도 질문 경로, 두 채팅 제어 분기를 소스와 비교합니다. JSON287개·다운로드5개·이벤트 스트림4개·텍스트1개입니다.
- 집중 테스트61개 통과: CLI25·런타임17·패키지13·업데이터6. 생략11개 중 패키지 E2E1개는 별도로 실행했습니다. 기존 업데이터10개는 Windows에서 Unix 셸·실행 파일 동작이 필요해 생략됐습니다.
- 설치된 패키지 E2E: 다른 임시 디렉터리에서 실제 PostgreSQL과 CLI 호출62회를 검증했습니다. 명령 탐색, 기능별 조회, 작업·목표·제약·회사·에이전트·대화, Unicode 파일, 바이너리 업로드·다운로드 무결성, Markdown·ZIP, 오류 종료 코드, 일시정지 저장, 재시작·데이터 이동을 포함합니다.
- 전체 폴더 이동: 패키지를 복사하고 C:\Windows\Temp 및 Windows 시스템 경로만 있는 PATH에서 기본 데이터 위치로 실행했습니다. Unicode·여러 줄 내용을 저장하고 종료한 다음 전체 폴더를 Unicode 경로로 옮겨 같은 내용을 읽었습니다. 종료 후 소유 프로세스0개, 원본 파일1,573개의 해시가 전후 일치했습니다. 최종 패키지에는 영·한 README와 CLI·검증 문서도 포함합니다.
- 런타임: 시작 요청4개가 같은 PID를 사용했고 클라이언트 종료 뒤에도 유지됐습니다. 인증 없는 health·초기화 요청은401로 거부했습니다. 기존 DB·설정·스킬 환경 변수로 다른 설치를 선택하지 않았습니다. 소유 런타임의 비정상 종료 뒤 인증된 남은 DB를 복구했습니다.
- 원본 DB 테스트336개와 에이전트 테스트534개가 실패·생략 없이 통과했습니다. 각각 새로운 scopeweaver_test_* 데이터베이스를 만들고 끝난 뒤 삭제했습니다.
- 독립 검토에서 Linux 업데이트 재실행 경로, 런타임 dry-run 변경 실행, REPL Windows 경로·표준입력 처리, 중단된 스킬 복사를 수정했고 재검토가 통과했습니다.
- 변경 패키지 go vet와 Linux·Darwin 런타임 교차 컴파일이 통과했습니다. 기존 WSL에서 Linux PostgreSQL17.11을 실행했고 macOS arm64 압축 해제를 확인했습니다.

나머지 엔진 패키지19개 중17개 통과·2개는 테스트가 없었습니다. 최상위 테스트291개(하위 포함545개)가 통과했습니다. 생략8개 중 DB 의존7개는 새로운 DB로 별도 재실행했고 나머지1개는 Windows 심볼릭 링크 권한 검사입니다.

## 기존 실패와 한계

추가 DB 의존 테스트는 evidence17개 통과·실패/생략0개, llmrec22개 통과·1개 실패·생략0개입니다. TestSideUsageRecordedOnceOnConsumerCancellation은 llm_usage 테이블 부재(SQLSTATE42P01)로 실패했으며 수정하지 않은 원본에서도 동일하게 재현됐습니다. 기존 실패는 총3개이며 이를 통과로 처리하지 않았습니다. 두 테스트 DB는 삭제했습니다. core-evidence-llmrec-evidence.json과 로컬 baseline-core-llmrec-test.jsonl에 근거가 있습니다.

원본 서버 테스트는491개 통과·2개 실패·4개 생략입니다. 두 실패는 수정하지 않은 원본 저장소(확인한 HEAD5681d41334804bb9411a85632bf5aefbb4f9cd70)에서도 재현됐습니다. CLI 작업으로 해당 소스를 바꾸거나 검증을 약화하지 않았습니다.

1. TestChatBranchCopyRejectsDestinationSymlink: Windows의 심볼릭 링크 생성 권한 부족.
2. TestChatStartupRecoveryFailureStopsConstruction: 대화 기록 경로가 파일인 상태에서 기존 생성자가 사용 가능한 서버를 반환함.

서버 생략 항목은 TestAgentAdapterE2E, TestChatControlsE2E, TestLiveContextReview, TestTaskArchivePackageSkipsSymlink입니다. 기존 서버 전체 테스트가 통과했다고 주장하지 않습니다.

깨끗한 Windows VM 및 Linux·macOS의 전체 CLI 실행 E2E는 수행하지 않았습니다. 현재 PC의 격리된 PATH 테스트는 원본 체크아웃·별도 앱·별도 DB 설치에 의존하지 않는 실행을 확인하지만 모든 OS 환경을 보증하지 않습니다. CI는 설정했으나 저장소를 공개하지 않아 호스팅 CI는 실행되지 않았습니다.

LLM·ScopeSentry·알림·사용자 도구·MCP는 설정된 외부 의존성입니다. 명령 대응과 엔진 테스트가 모든 외부 서비스를 실제 호출했다는 증거는 아닙니다. CLI 안정 배포가 아직 없어 온라인 업데이트 다운로드는 검증하지 않았으며 업데이터 단위 검증·재실행 경로를 확인했습니다.

## 근거와 재현

[근거 폴더](../evidence/standalone-cli/)에 E2E 로그·런타임·전체 이동 기록·핵심 테스트 수·독립 검토 기록을 보존합니다. 자세한 로컬 로그는 dist/core-*-test.jsonl, dist/baseline-core-server-test.jsonl, dist/focused-tests.jsonl에 있습니다.

    go test ./internal/cli ./internal/portable ./packaging ./selfupdate -count=1
    go vet ./internal/cli ./internal/portable ./packaging ./selfupdate
    go run ./packaging -target windows-amd64
    # SCOPEWEAVER_CLI_E2E_BINARY를 패키지 실행 파일로 지정하세요.
    go test ./internal/cli -run '^TestPortableCLIEndToEnd$' -count=1 -v

이 빌드를 설명할 때 위 한계를 함께 전달하세요.

제3자 라이선스는 공백을 포함한 게시자 원본 바이트와 해시를 보존합니다. 전체 Git 공백 검사는 해당 원본 파일을 지적하지만 자체 소스 공백 검사는 통과했습니다. 고정된 라이선스 파일에는 Git 줄바꿈 변환을 적용하지 않습니다.
