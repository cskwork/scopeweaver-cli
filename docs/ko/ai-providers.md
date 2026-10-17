# AI 제공자 설정과 한계

[English](../ai-providers.md) · [CLI 구조](cli.md) · [검증](CLI-VERIFICATION.md)

내장 채팅·플래너·워커 에이전트는 LLM 프로필 형식으로 `codex-cli`를 사용할 수 있습니다. 어댑터는 텍스트와 도구 설명을 표준 입출력으로 공식 Codex app-server에 보냅니다. 엔진의 에이전트 실행 계층인 Norma가 요청된 ScopeWeaver 도구를 실행하고 그 결과를 다음 모델 요청에 전달합니다. Codex에 도구 실행을 맡기지 않습니다.

## ChatGPT 로그인

ChatGPT 계정으로 로그인한 공식 Codex CLI가 필요합니다. 검증한 버전은 Codex 0.162.1이며 다른 버전이나 향후 버전의 호환성은 확인하지 않았습니다. Codex는 별도 의존성이며 이식 가능한 PostgreSQL 실행 환경을 준비할 때 내려받지 않습니다.

```sh
codex --version
codex login
codex login status
```

브라우저 로그인은 공식 Codex가 처리합니다. ScopeWeaver가 OAuth 토큰을 직접 추출·복사·전송하지 않습니다. 이 어댑터의 ChatGPT 경로는 OpenAI API 키 로그인을 허용하지 않습니다. 모델 접근 권한과 사용 한도는 로그인한 계정에 따릅니다. [공식 인증 안내](https://learn.chatgpt.com/docs/auth)를 확인하세요.

Codex는 PATH에서 찾습니다. 또는 이식 가능한 런타임을 시작하기 전에 `SCOPEWEAVER_CODEX_PATH`를 네이티브 실행 파일 경로로 지정하세요. Windows에서는 `.cmd` 래퍼가 아닌 `.exe`를 지정해야 합니다. 공식 npm 패키지로 설치한 경우 어댑터가 Node.js 실행 없이 Windows 네이티브 실행 파일을 찾을 수 있습니다.

다음을 `chatgpt-profile.json`으로 저장하세요.

```json
{
  "name": "ChatGPT via Codex",
  "format": "codex-cli",
  "model": "gpt-6.1-sol",
  "base_url": "",
  "api_key": "",
  "reasoning_effort": "high",
  "streaming": true
}
```

```sh
scopeweaver-cli schema llm profiles save --json
scopeweaver-cli llm profiles save --data @chatgpt-profile.json --yes --json
scopeweaver-cli llm profiles list --json
scopeweaver-cli llm profiles activate --id 1 --yes --json
scopeweaver-cli llm test --provider codex-cli --model gpt-6.1-sol --profile-id 1 --yes --json
```

`1`을 저장 응답의 `id`로 바꾸세요. 저장은 설정만 검증하며 Codex를 시작하거나 모델에 접속하지 않습니다. `llm test`는 실제 모델 요청을 보냅니다. 이 경로에서는 `base_url`, `api_key`, `proxy`를 비워 두세요. 프로필을 활성화하면 기본값으로 사용하며 작업·에이전트에 명시적으로 지정한 프로필의 우선순위는 유지됩니다.

## Z.AI GLM Coding Plan

어댑터는 고정된 Z.AI Responses 주소에서 `glm-5.3`과 `glm-5.3-flash`를 지원합니다. 임시로 만든 깨끗한 Codex 홈, 제공자 전용 인증 정보와 포함된 두 GLM 모델의 정보로 공식 Codex를 시작합니다. 사용자의 Codex 설정을 바꾸거나 Z.AI 요청에 ChatGPT 로그인을 사용하지 않습니다.

이식 가능한 런타임을 시작하는 프로세스의 환경에 `ZAI_API_KEY`를 제공하세요. 런타임이 이미 실행 중이면 `runtime stop --yes`로 중지하고 환경 변수를 안전하게 설정한 뒤 다음 기능 명령을 실행하세요. 또는 파일 탐색이 가능한 작업 공간 밖의 비공개 프로필 JSON 파일에 `api_key`를 지정할 수 있습니다. 인증 정보를 커밋하거나 검증 자료에 넣지 마세요.

`ZAI_API_KEY`를 사용할 때는 키가 없는 다음 설정을 `zai-profile.json`으로 저장하세요.

```json
{
  "name": "Z.AI Coding Plan via Codex",
  "format": "codex-cli",
  "model": "glm-5.3",
  "base_url": "https://api.z.ai/api/v1",
  "reasoning_effort": "max",
  "streaming": true,
  "context_window_k": 1000
}
```

```sh
scopeweaver-cli llm profiles save --data @zai-profile.json --yes --json
scopeweaver-cli llm profiles activate --id 2 --yes --json
scopeweaver-cli llm test --provider codex-cli --model glm-5.3 --base-url https://api.z.ai/api/v1 --profile-id 2 --reasoning-effort max --yes --json
```

`2`를 저장 응답의 프로필 ID로 바꾸세요. `/responses`로 끝나는 주소는 같은 루트 주소로 정규화합니다. Flash를 사용하려면 같은 프로필의 모델을 `glm-5.3-flash`로 바꾸고 `llm test`에도 같은 모델을 지정하세요. 다른 주소나 Z.AI 모델은 거부합니다. 두 모델의 추론 강도는 `low`, `high`, `max`를 지원하며 사고 비활성화 요청은 `low`로 처리합니다.

[Z.AI 공식 Codex 안내](https://docs.z.ai/devpack/tool/codex)를 따르고 [API 키 관리 안내](https://docs.z.ai/devpack/notice/api-key-management)에서 키 유형을 확인하세요. Standard·Coding Plan·Compatible 키는 사용 규칙이 다르며 Standard 키는 Coding Plan 할당량을 사용하지 않습니다. 사용자는 제공한 키가 Personal Coding Plan·Team Plan 키라고 확인했습니다. 첫 번째 키의 호출은 insufficient_quota를 반환했고 내장 제공자 직접 호출도 usageLimitExceeded와 토큰0을 보고했습니다. 이후 교체된 키로 내장 GLM-5.3 추론과 공식 Codex의 GLM-5.3-Flash 호출이 성공했습니다. Windows R4 GLM-5.3-Flash/low는 실제 CLI 호출 127회를 153.87초에 통과했으며 채팅·도구 결과, 대기열·재개, 플래너·워커, 소유한 루프백 GET 1회, 발견 사항·목표 완료·보고서를 확인했습니다. GLM-5.3은 채팅·인텐트 생성·워커 HTTP 근거 수집·발견 사항 등록까지 완료했지만 이후 목표 확인 호출이 E2E 제한 시간을 넘겼습니다. 이 실패를 Flash의 통과와 구분하며, 대표 모델 GLM-5.3의 전체 흐름 통과로 주장하지 않습니다. 이 결과로 이전 계정 오류의 원인까지 단정할 수는 없습니다. 전체 워크플로 검증은 날짜별 검증 기록을 참조하세요.

## 동작과 호환성 한계

- 모델 요청마다 소유한 app-server 프로세스와 새 임시 스레드를 만듭니다. 현재 Norma 대화 전체의 역할, 도구 ID·결과와 사용 가능한 도구 스키마를 매번 JSON으로 변환합니다. 대화 기록·분기·재시작 후 보존은 ScopeWeaver 엔진이 담당하며 Codex 스레드를 그대로 재생하는 방식은 아닙니다.
- 스트리밍은 완료된 구조화 응답의 검증이 끝날 때까지 버퍼링합니다. 이후 텍스트·도구 이벤트를 기존 에이전트 스트림으로 전달합니다. 완료 전 실시간 토큰 델타는 없으며 불완전하거나 잘못된 도구 요청은 Norma에 전달하지 않습니다.
- 이 요청에서는 Codex의 셸·웹·이미지·MCP·플러그인·앱·스킬·훅·하위 에이전트를 비활성화합니다. 추론 전에 MCP 격리를 확인하며 지원하지 않는 MCP 이름은 명시적으로 실패합니다. 예상하지 않은 도구 항목이나 app-server 요청이 나오면 모델 요청을 거부합니다. ScopeWeaver에서 설정한 도구·MCP 연동은 기존 제어에 따라 Norma가 실행합니다.
- 프로젝트 지침과 환경 접근을 비활성화합니다. ChatGPT OAuth는 로그인에 공식 Codex 홈을 사용하므로 전역 `AGENTS.md`를 포함한 전역 지침을 상속합니다. Z.AI는 임시로 만든 깨끗한 홈을 사용합니다. 따라서 계정의 전역 지침이 OAuth 경로의 응답에 영향을 줄 수 있습니다.
- 텍스트·사고·구조화된 도구 기록을 지원하며 이미지와 다른 비텍스트 입력은 거부합니다. 변환된 문맥은 32 MiB, 구조화 출력은 2 MiB, 도구 호출은 요청당 128개로 제한합니다. 각 호출의 제한 시간은 10분이며 호출자가 먼저 취소할 수 있습니다.
- `max_tokens`, 온도·중지 문자열·사고 비활성화 요청은 모델에 주는 참고 지침이며 API가 강제하는 제어가 아닙니다. OAuth 문맥 압축은 지원하지 않는 `none`을 강제하지 않고 프로필의 추론 강도를 유지합니다. GLM의 사고 비활성화 요청은 `low`로 처리합니다. 프로필의 문맥 크기는 엔진의 문맥 압축 기준이며 제공자의 실제 처리 용량을 바꾸지 않습니다. OAuth 추론 강도는 빈 값, `none`, `minimal`, `low`, `medium`, `high`, `xhigh`를 허용하지만 선택한 모델이 더 제한할 수 있습니다.
- 프로필 프록시는 거부합니다. 사용자 지정 세션 헤더, HTTP 제공자 재시도 옵션과 원시 HTTP 요청·응답 수집은 지원하지 않습니다. 설정된 엔진의 재시도·대체 제공자·요청 빈도·동시성 제어는 적용되는 경로에서 유지됩니다. 사용량은 app-server 토큰 사용 알림에서 읽으며 원시 Codex 진단은 노출하지 않습니다.
- 두 경로 모두 클라우드 제공자입니다. 매 호출마다 `local_llm_only`를 확인하고 활성화되어 있으면 Codex 시작 전에 차단합니다. 허용 CIDR을 추가해도 로컬 추론으로 바뀌지 않습니다. 프로필 저장이나 설정 준비 완료만으로 실행 파일·로그인·권한·네트워크 준비가 검증되지는 않습니다.
- `codex-cli`는 `llm models list`를 통한 모델 탐색을 지원하지 않습니다. 정확한 모델을 직접 설정하세요. 패키지 준비와 핵심 기능의 오프라인 재사용이 클라우드 추론까지 오프라인으로 바꾸지는 않습니다.

## 검증 범위

Windows R3 OAuth는 설치형 CLI 호출 125회를 161.24초에 통과했습니다. 별도 수명주기 검증은 호출 130회와 사전 점검 3회로 실행 중 채팅·작업 및 보조 질문의 종료 대기·재시작, 진행 방향 지시, 편집·재시도·분기를 확인했습니다. 보조 질문 취소가 보조 추론 완료를 증명하지는 않습니다. Windows R4 Flash는 위 내장 흐름을 통과했습니다. 특정 후보의 결과이며 모든 모델·플랫폼이나 최종 패키지의 통과를 뜻하지 않습니다. [날짜별 기록과 남은 조건](CLI-VERIFICATION.md).

R4 OAuth 문맥 압축은 CLI 호출 33회와 새 모델 요청 2회로 통과했습니다. 요청별 유료 기록과 요약은 재시작 후 유지됐지만 기존 대화·대시보드 집계 사용량은 늘지 않았습니다. 압축 비용이 해당 집계에 포함됐다고 확인한 것은 아닙니다. 최신 최종 출력 오류·다국어 수정은 소스 테스트 근거만 있으므로 재빌드한 후보 검증이 필요합니다.

소스 저장소의 구현 근거: `codexprovider/provider.go`, `codexprovider/protocol.go`, `codexprovider/isolation.go`, `agent/provider.go`, `internal/cli/operations.json`. 공식 [app-server 문서](https://learn.chatgpt.com/docs/app-server)는 버전별 프로토콜을 설명합니다. 이 어댑터의 검증 기준은 0.162.1입니다.
