# 포터블 패키지 만들기

[English](README.md)

Go 1.26.3 이상으로 저장소 최상위 폴더에서 실행합니다.

```powershell
go run ./packaging -target windows-amd64
```

`dist/scopeweaver-cli-windows-amd64`에 CLI 실행 파일, 전용 PostgreSQL 17.11,
스킬, 라이선스, SHA-256 매니페스트가 생성됩니다. 이 폴더 전체를 복사하면 됩니다.
사용자 PC에 Go, Node, Docker, PostgreSQL 또는 PostgreSQL 서비스를 별도로
설치하지 않습니다. 패키징 도구는 프로그램 설치나 데이터베이스 변경을 수행하지
않습니다. CLI에서 선택적으로 사용하는 외부 도구와 LLM 공급자는 별도 설정이
필요합니다.

대상은 `windows-amd64`, `linux-amd64`, `linux-arm64`, `darwin-amd64`,
`darwin-arm64`입니다. Linux는 glibc용이며 Alpine/musl용이 아닙니다.
Windows 10 이상이 필요합니다. macOS와 Linux의 기본 시스템 라이브러리는 계속
사용합니다. 압축 해제나 교차 컴파일에 성공했다고 해당 운영체제에서 실행을 검증한
것은 아닙니다.

```powershell
# 출력 폴더는 새로 생성하며 기존 폴더를 덮어쓰지 않습니다.
go run ./packaging -target windows-amd64 -out dist/my-portable-folder

# 고정된 다운로드가 캐시에 있으면 네트워크 없이 다시 빌드합니다.
go run ./packaging -target windows-amd64 -offline -out dist/offline-build

# CLI 개발 완료 전에 런타임을 준비합니다.
go run ./packaging -target windows-amd64 -runtime-only -out dist/runtime-only

go test ./packaging -count=1
```

`-runtime-only`는 개발용이며 CLI 실행 파일을 포함하지 않습니다.
다운로드는 `dist/cache`에 보관합니다. 캐시의 해시가 다르면 교체하지 않고 실패합니다.
실패한 출력 폴더는 점검을 위해 남기므로 재시도할 때는 새 경로를 지정합니다.

`postgres-sources.json`은 배포 파일과 SHA-256을 고정합니다.
배포자는 [Zonky Embedded Postgres Binaries](https://github.com/zonkyio/embedded-postgres-binaries)이며
Windows/macOS에서는 EDB 바이너리를 재포장하고 Linux용 PostgreSQL은 직접 빌드합니다.
원래 임베디드 테스트용 번들이며 이 CLI는 사용자별 로컬 데이터베이스에 사용합니다.
[PostgreSQL Windows 안내](https://www.postgresql.org/download/windows/)와
[EDB 바이너리 안내](https://www.enterprisedb.com/download-postgresql-binaries)에서
바이너리 배포 정보를 확인할 수 있습니다. 축소 번들은 `initdb`, `pg_ctl`,
`postgres`를 포함하며 `psql`은 포함하지 않습니다.

Windows에는 Microsoft VCLibs에서 추출한 Visual C++ DLL 3개도 실행 파일 옆에
포함합니다. 체크섬은 [Microsoft winget 매니페스트](https://github.com/microsoft/winget-pkgs/blob/master/manifests/m/Microsoft/VCLibs/Desktop/14/14.0.33321.0/Microsoft.VCLibs.Desktop.14.installer.yaml)와
대조했습니다. 런타임 설치 프로그램은 실행하지 않습니다. 출처 URL, 해시,
Microsoft 원문 라이선스가 결과물에 포함됩니다.
[Microsoft 배포 안내](https://learn.microsoft.com/en-us/cpp/windows/redistributing-visual-cpp-files)는
이 방식의 배포를 설명하며 재배포 권한을 Visual Studio 라이선스 사용자로 제한합니다.
**외부 공개 전에는 Microsoft 재배포 권한과 포함된 LGPL 라이브러리의 소스 제공
의무를 충족해야 합니다.** 로컬 패키징 성공이 제삼자 바이너리의 공개 권한을
뜻하지는 않습니다.

압축 해제는 Go의 루트 제한 파일 API로 수행합니다. 절대 경로, 외부를 가리키는 링크,
Windows 예약 장치 이름, 중복 파일, 장치 파일은 거부합니다. 내부 라이브러리 링크는
일반 파일로 복사하므로 Windows에서도 심볼릭 링크 권한 없이 다른 운영체제용
패키지를 만들 수 있습니다. `github.com/ulikunitz/xz`는 패키징 도구의 빌드에만
사용하며 배포 CLI의 의존성은 아닙니다.

`MANIFEST.json`은 소스 리비전, 미커밋 변경 여부, 대상, PostgreSQL 출처 URL/해시와
매니페스트 자신을 제외한 모든 일반 파일의 SHA-256을 기록합니다.
`POSTGRES-SOURCES.json`에는 PostgreSQL과 Windows 런타임의 고정 출처가 있습니다.
`licenses/SOURCES.json`은 원문 고지의 출처와 해시를 기록합니다.
Go 라이선스 폴더에는 CLI에 연결된 모듈과 표준 라이브러리의 라이선스를 보관합니다.
패키지를 옮길 때 라이선스 파일도 함께 보관해야 합니다.
