# scopeweaver-cli

ScopeWeaver 기능을 에이전트가 명령줄에서 사용할 수 있는 독립 실행형 CLI입니다. 배포 폴더에 Go 엔진, PostgreSQL 실행 파일, 기본 스킬과 의존성 고지가 포함됩니다. 별도의 ScopeWeaver·데이터베이스·Docker·Node.js·Go 설치 없이 실행합니다.

[English](README.md) · [구조와 호환성](docs/ko/cli.md) · [검증](docs/ko/CLI-VERIFICATION.md)

## 시작

PowerShell에서는 현재 폴더의 실행 파일을 다음과 같이 실행합니다.

```powershell
.\scopeweaver-cli.exe doctor --json
.\scopeweaver-cli.exe commands --json
.\scopeweaver-cli.exe schema tasks create --json
.\scopeweaver-cli.exe health --lang ko --json
.\scopeweaver-cli.exe tasks list --json
.\scopeweaver-cli.exe runtime stop --yes --json
```

처음 기능 명령을 실행하면 전용 로컬 저장소를 초기화하고 포함된 엔진과 데이터베이스를 시작합니다. 이후 명령은 같은 런타임을 사용하므로 명령이 끝나도 진행 중인 작업은 계속됩니다. 운영체제 서비스는 등록하지 않습니다.

기본 데이터 위치는 실행 파일 옆 `state/`입니다. `--home PATH` 또는 `SCOPEWEAVER_CLI_HOME`으로 작업 공간을 분리할 수 있습니다. 폴더를 이동하거나 복사하기 전에 런타임을 중지하세요. 다른 운영체제나 PostgreSQL 주 버전으로 옮길 때는 지원되는 백업·복원이 필요합니다. 데이터베이스 디렉터리를 그대로 복사하는 방식은 이식용 교환 형식이 아닙니다.

## 에이전트 사용

`schema`는 명령별 매개변수 형식, 필수 값, 실행 영향, 응답 형식과 근거 소스 위치를 출력합니다. 모든 백엔드 경로와 확장된 채팅 제어를 이름 있는 명령으로 제공합니다.

```sh
scopeweaver-cli tasks create --data @task.json --dry-run --json
scopeweaver-cli tasks create --data @task.json --yes --json
scopeweaver-cli tasks get --id 1 --json
scopeweaver-cli tasks control --id 1 --action pause --yes --json
scopeweaver-cli workspace upload --path fixtures --upload example.txt --yes --json
scopeweaver-cli workspace download --path fixtures/example.txt --output downloaded.txt --json
scopeweaver-cli logs watch --json
```

`task.json` 예시:

```json
{
  "name": "격리된 테스트 자료",
  "description": "로컬 격리 환경의 가상 자료만 사용합니다. 온라인 대상에 접속하지 않습니다.",
  "goal": "승인된 가상 자료를 확인합니다."
}
```

- 조회는 바로 실행됩니다. 상태 변경, 도구 실행, 외부 연동 호출, 알림 전송, 업데이트는 `--yes`가 필요합니다.
- `--dry-run`은 런타임을 시작하지 않고 요청을 검증·출력합니다. 알려진 인증 정보 필드는 가립니다.
- `--data @FILE`은 JSON 객체 파일, `--data -`는 표준 입력을 읽습니다. 이름 있는 옵션과 반복 가능한 `--field key=value`로 일부 필드를 지정할 수 있습니다. 중첩 입력과 인증 정보는 JSON 파일을 사용하세요.
- `--query key=value`를 반복해 백엔드 조회 옵션을 전달합니다. 매개변수 옵션의 밑줄과 하이픈은 서로 바꿔 쓸 수 있습니다.
- 다운로드에는 `--output`이 필요하며 기존 파일을 덮어쓰지 않습니다. 스트림은 한 줄에 JSON 이벤트 하나를 출력하고 Ctrl+C로 중지합니다.
- JSON 결과는 표준 출력, 오류·진단은 표준 오류로 전달합니다. 종료 코드: `0` 성공, `1` 작업 실패, `2` 잘못된 입력 또는 확인 필요, `3` 런타임 사용 불가, `4` 인증·권한 오류, `130` 취소.
- 작업 생성 성공은 접수 성공입니다. 완료 여부는 상태·활동·발견 사항으로 확인하세요. 결과가 불확실한 변경 요청은 자동 재시도하지 않습니다.
- `repl`에서 같은 명령을 대화형으로 실행합니다. `exit`는 셸을 닫고, `runtime stop --yes`는 백그라운드 작업을 중지합니다. 셸 명령 치환은 실행하지 않습니다.

코딩 에이전트에서 [스킬](skills/scopeweaver-cli/SKILL.md)을 읽거나 `scopeweaver-cli schema --json`으로 기능을 확인하세요. 명령 이름과 JSON 키는 고정이며 `--lang en|ko|zh-Hans`로 화면·백엔드 메시지 언어를 선택합니다. 원본 증거와 기술 진단은 보존합니다.

## 기능 범위

작업·워커, 목표·제약·범위, 분류·템플릿·보관, 회사·자산·동기화, 탐색·커버리지 그래프, 발견 사항·재검증·트래픽 증거·보고서, 작업 공간 파일, 대화·채팅 제어·별도 질문, 에이전트·프롬프트·트리거, 도구·스킬·MCP, LLM 프로필·재시도·풀·기록, 승인·차단, 알림, 설정, 통계·토큰·로그, 정리와 업데이트를 다룹니다.

그래프는 노드·간선 데이터로 제공합니다. 브라우저 화면 배치, 테마와 키보드 설정은 CLI 범위에 포함하지 않습니다.

## 모델과 연동 요구 사항

AI 실행에는 모델 제공자 또는 로컬 추론 서비스 설정이 필요합니다. 모델 가중치, 유료 구독과 API 키는 포함하지 않습니다. 사용자 지정 명령, Python 도구, 브라우저 도구, 외부 MCP·서비스는 해당 연동에서 선택한 실행 파일이나 서비스를 요구합니다. `doctor`는 포함된 핵심 실행 환경과 선택적 연동 요구 사항을 구분합니다.

로컬 런타임은 인증을 요구하며 루프백 주소만 사용합니다. 전용 인증 정보는 파일 탐색이 가능한 작업 공간 밖에 둡니다. 증거와 도구 출력은 신뢰할 수 없는 데이터로 취급하세요.

업데이트는 `cskwork/scopeweaver-cli` 배포를 대상으로 CLI·엔진 실행 파일을 교체합니다. 데이터와 사용자 수정 스킬은 보존하며 PostgreSQL 주 버전을 자동 변경하지 않습니다. 배포 전에는 안정 버전이 없다는 오류를 표시합니다.

## 빌드와 검증

개발에는 Go 1.26.3과 고정된 의존성·실행 파일을 내려받을 네트워크 연결이 필요합니다. 배포본 사용자에게는 필요하지 않습니다.

```sh
go test ./internal/cli ./internal/portable ./packaging ./selfupdate
go run ./packaging -target windows-amd64
```

패키지 대상은 `windows-amd64`, `linux-amd64`, `linux-arm64`, `darwin-amd64`, `darwin-arm64`입니다. 대상별 PostgreSQL 실행 파일을 사용하고 고정 SHA-256 및 파일별 명세를 검증합니다. `-offline`은 검증된 캐시만 사용합니다. 실행 시에는 의존성을 내려받지 않습니다. Linux PostgreSQL은 root가 아닌 사용자로 실행해야 합니다.

설치된 실행 파일 E2E:

```powershell
$env:SCOPEWEAVER_CLI_E2E_BINARY = (Resolve-Path .\dist\scopeweaver-cli-windows-amd64\scopeweaver-cli.exe).Path
go test ./internal/cli -run '^TestPortableCLIEndToEnd$' -count=1 -v
```

E2E는 임시 로컬 상태를 만들고 실제 PostgreSQL, 저장된 변경, 업로드·다운로드, 재시작, 다른 작업 디렉터리에서의 실행과 데이터 이동을 확인합니다. 실제 대상이나 모델 제공자 인증 정보는 사용하지 않습니다.

## 출처와 사용 범위

ScopeWeaver 커밋 `2df0b307dcb817a459abebd3a4c240a2019598db`를 기반으로 하며 원본은 ARTEX에서 파생됐습니다. 추적을 위해 원본 소스와 테스트를 보존합니다. 배포 빌드는 Go 엔진만 사용하고 웹 UI를 빌드하지 않습니다. 명령 탐색·구조화 출력·실제 앱 검증 방식은 [CLI-Anything](https://github.com/HKUDS/CLI-Anything)을 참고했습니다.

[AGPL-3.0](LICENSE), [제3자 고지](THIRD_PARTY_NOTICES.txt), 상위 프로젝트 사용 제한이 유지됩니다. 소스 연구와 로컬 격리 환경의 기술 검증에만 사용하세요. 소유하거나 허가받은 대상도 포함해 온라인 시스템에 대한 실제 테스트는 상위 프로젝트에서 금지합니다. [보존된 출처 문서](docs/ko/PROVENANCE.md)를 확인하세요.

동봉된 실행 환경의 라이선스는 `licenses/`, 소스 정보는 `POSTGRES-SOURCES.json`에 있습니다. 패키지 생성은 공개 배포를 뜻하지 않습니다.
