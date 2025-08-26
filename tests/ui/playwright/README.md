## installation
pip install -r requirements.txt
playwright install # install playwright tools and browsers


## record new test
playwright codegen https://stork.lab.isc.org/
playwright codegen https://stork.lab.isc.org/ --output test_example.py --target python-pytest

# generated code copy into tests/ui/playwright/test_basic.py 


## start environment
docker compose -p stork_tests -f tests/system/docker-compose.yaml -f tests/system/docker-compose.ui.yaml build server
docker compose -p stork_tests -f tests/system/docker-compose.yaml -f tests/system/docker-compose.ui.yaml up -d --no-build postgres server agent-kea



## run tests
pytest --headed --base-url=https://stork.lab.isc.org/ --browser chromium --browser firefox --tracing on <path_to_test_file>
# options can be added to pytest.ini 


## debug failing tests
1.playwright show-trace test-results/<test_name>/trace.zip
2.PWDEBUG=1 pytest tests/ui/playwright/<test_name>.py

