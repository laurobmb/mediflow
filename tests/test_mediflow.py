import os
import unittest
import time
import logging
import subprocess
import re
import random
import string
from datetime import datetime, timedelta
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.support.ui import WebDriverWait, Select
from selenium.webdriver.support import expected_conditions as EC
from webdriver_manager.chrome import ChromeDriverManager
from selenium.common.exceptions import TimeoutException

# --- Configurações ---
logging.basicConfig(format='%(asctime)s: %(name)s: %(levelname)s: %(message)s', level=logging.INFO, datefmt='%H:%M:%S')
logger = logging.getLogger('MediFlowSystemTest')

BASE_URL = os.getenv('APP_URL', 'http://127.0.0.1:8080')
DEFAULT_TIMEOUT = 20 # Timeout padrão para a maioria das ações
AI_TIMEOUT = 120 # Timeout estendido para aguardar a resposta da IA
STEP_DELAY = 0.5

# --- Credenciais de Teste ---
ADMIN_EMAIL = "admin@mediflow.com"
TEST_PASS = "senha123"
TEST_CPF = "34657488082"
TERAPEUTA_NOME = "Dr. Exemplo"

class MediFlowSystemTest(unittest.TestCase):
    
    @classmethod
    def setUpClass(cls):
        """Configura o ambiente uma vez antes de todos os testes."""
        logger.info("A preparar o ambiente de teste...")

        # --- LÓGICA PARA SCREENSHOTS ---
        cls.screenshot_dir = 'photos'
        os.makedirs(cls.screenshot_dir, exist_ok=True)
        cls.screenshot_counter = 0
        logger.info(f"Diretório de screenshots '{cls.screenshot_dir}' garantido.")
        
        # --- 1. CRIAR BANCO DE DADOS DE TESTE ---
        logger.info("A criar o banco de dados de teste...")
        try:
            result = subprocess.run(
                ['go', 'run', 'data_manager.go', '-create-test-db'],
                capture_output=True, text=True, check=True
            )
            cls.test_db_name = result.stdout.strip().split('\n')[-1]
            logger.info(f"Banco de dados de teste '{cls.test_db_name}' criado com sucesso.")
        except subprocess.CalledProcessError as e:
            logger.error(f"Falha ao criar o banco de dados de teste: {e.stderr}")
            exit(1)

        # --- 2. POPULAR O BANCO DE DADOS DE TESTE ---
        logger.info(f"A popular o banco de dados '{cls.test_db_name}'...")
        try:
            subprocess.run(
                ['go', 'run', 'data_manager.go', '-init', '-create-users', '-populate', f'-dbname={cls.test_db_name}'],
                capture_output=True, text=True, check=True
            )
            logger.info("Banco de dados populado com sucesso.")
        except subprocess.CalledProcessError as e:
            logger.error(f"Falha ao popular o banco de dados de teste: {e.stderr}")
            exit(1)

        # --- 3. INICIAR A APLICAÇÃO MEDIFLOW ---
        logger.info("A iniciar a aplicação MediFlow...")
        app_env = os.environ.copy()
        app_env["DB_NAME"] = cls.test_db_name
        
        cls.mediflow_process = subprocess.Popen(
            ['go', 'run', 'main.go'],
            env=app_env,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True
        )
        time.sleep(5) 
        logger.info("Aplicação MediFlow iniciada.")

        # --- 4. INICIAR O NAVEGADOR ---
        options = webdriver.ChromeOptions()
        options.add_argument("--start-maximized")
        try:
            service = Service(ChromeDriverManager().install())
            cls.browser = webdriver.Chrome(service=service, options=options)
            cls.wait = WebDriverWait(cls.browser, DEFAULT_TIMEOUT)
            logger.info("ChromeDriver iniciado com sucesso.")
        except Exception as e:
            logger.error(f"Não foi possível iniciar o ChromeDriver: {e}")
            cls.mediflow_process.terminate()
            exit(1)

    @classmethod
    def tearDownClass(cls):
        """Encerra tudo no final dos testes."""
        if cls.browser:
            cls.browser.quit()
            logger.info("Navegador encerrado.")
        if hasattr(cls, 'mediflow_process') and cls.mediflow_process:
            cls.mediflow_process.terminate()
            logger.info("Aplicação MediFlow encerrada.")
        if hasattr(cls, 'test_db_name'):
            logger.info(f"O banco de dados de teste '{cls.test_db_name}' NÃO foi removido para inspeção.")

    def _login(self, email, password):
        """Função auxiliar para fazer login."""
        logger.info(f"A tentar fazer login como {email}...")
        self.browser.get(f'{BASE_URL}/login')
        self.wait.until(EC.presence_of_element_located((By.ID, "email"))).send_keys(email)
        self.browser.find_element(By.ID, "password").send_keys(password)
        self.browser.find_element(By.CSS_SELECTOR, "button[type='submit']").click()
        time.sleep(STEP_DELAY)

    def _logout(self):
        """Função auxiliar para fazer logout."""
        logger.info("A fazer logout...")
        self.wait.until(EC.element_to_be_clickable((By.XPATH, "//a[contains(@href, '/logout')]"))).click()
        self.wait.until(EC.url_contains('/login'))
        self._take_screenshot("logout_final")
        logger.info("Logout bem-sucedido.")

    def _take_screenshot(self, name):
        """Tira e salva uma screenshot numerada."""
        self.__class__.screenshot_counter += 1
        filename = f"{self.__class__.screenshot_counter:02d}_{name}.png"
        filepath = os.path.join(self.__class__.screenshot_dir, filename)
        try:
            self.browser.save_screenshot(filepath)
            logger.info(f"Screenshot salva em: {filepath}")
        except Exception as e:
            logger.error(f"Falha ao salvar screenshot {filepath}: {e}")

    def test_admin_full_flow(self):
        """Testa o fluxo completo do administrador: navegação, gestão de usuários e pacientes."""
        logger.info("--- INICIANDO TESTE COMPLETO DO FLUXO DE ADMINISTRADOR ---")
        self._login(ADMIN_EMAIL, TEST_PASS)
        self.wait.until(EC.url_contains('/admin/dashboard'))
        self._take_screenshot("dashboard_inicial")

        # 1. Teste de Navegação (Smoke Test)
        logger.info("Admin: Verificando todas as abas de navegação...")
        self.browser.get(f'{BASE_URL}/admin/agenda')
        self.wait.until(EC.url_contains("/admin/agenda"))
        self._take_screenshot("agenda")
        
        self.browser.get(f'{BASE_URL}/admin/users')
        self.wait.until(EC.url_contains("/admin/users"))
        self._take_screenshot("gerenciar_usuarios")
        
        self.browser.get(f'{BASE_URL}/admin/patients')
        self.wait.until(EC.url_contains("/admin/patients"))
        self._take_screenshot("gerenciar_pacientes")
        
        self.browser.get(f'{BASE_URL}/admin/monitoring')
        self.wait.until(EC.url_contains("/admin/monitoring"))
        self._take_screenshot("monitoramento")

        self.browser.get(f'{BASE_URL}/admin/audit-logs')
        self.wait.until(EC.url_contains("/admin/audit-logs"))
        self._take_screenshot("logs_auditoria")

        logger.info("Admin: Todas as abas foram acessadas com sucesso.")

        # 2. Gestão de Usuários
        random_str = ''.join(random.choices(string.ascii_lowercase, k=6))
        novo_user_nome = f"Teste User {random_str}"
        novo_user_email = f"teste.user.{random_str}@email.com"
        
        logger.info(f"Admin: A adicionar o utilizador '{novo_user_nome}'...")
        self.browser.get(f'{BASE_URL}/admin/users/new')
        self.wait.until(EC.presence_of_element_located((By.ID, "name"))).send_keys(novo_user_nome)
        self.browser.find_element(By.ID, "email").send_keys(novo_user_email)
        self.browser.find_element(By.ID, "password").send_keys("senha123")
        Select(self.browser.find_element(By.ID, "user_type")).select_by_visible_text("secretaria")
        self._take_screenshot("formulario_novo_usuario")
        self.browser.find_element(By.CLASS_NAME, "btn-submit").click()
        
        logger.info(f"Admin: A remover o utilizador '{novo_user_nome}'...")
        linha_user = self.wait.until(EC.presence_of_element_located((By.XPATH, f"//td[text()='{novo_user_nome}']/parent::tr")))
        self._take_screenshot("lista_com_novo_usuario")
        linha_user.find_element(By.LINK_TEXT, "Remover").click()
        self.wait.until(EC.alert_is_present()).accept()
        self.wait.until(EC.staleness_of(linha_user))
        self._take_screenshot("lista_apos_remocao_usuario")
        logger.info("Admin: Utilizador removido com sucesso.")

        # 3. Gestão de Pacientes e Fluxo de Consentimento
        random_str_patient = ''.join(random.choices(string.ascii_lowercase, k=6))
        novo_paciente_nome = f"Paciente Completo {random_str_patient}"
        logger.info(f"Admin: A adicionar o paciente '{novo_paciente_nome}' com dados completos...")
        self.browser.get(f'{BASE_URL}/admin/patients/new')
        
        # Preenche todos os campos do formulário
        self.wait.until(EC.presence_of_element_located((By.ID, "client_name"))).send_keys(novo_paciente_nome)
        self.browser.find_element(By.NAME, "address_street").send_keys("Rua dos Testes, 123")
        self.browser.find_element(By.NAME, "mobile").send_keys("11912345678")
        self.browser.find_element(By.NAME, "dob").send_keys("1990-10-15") # Formato YYYY-MM-DD
        self.browser.find_element(By.NAME, "email").send_keys(f"paciente.{random_str_patient}@teste.com")
        self.browser.find_element(By.NAME, "profession").send_keys("Engenheiro de Testes")
        
        logger.info("Admin: Preenchendo avaliação emocional...")
        self.browser.find_element(By.XPATH, "//input[@name='anxiety_level'][@value='8']").click()
        self.browser.find_element(By.XPATH, "//input[@name='anger_level'][@value='3']").click()
        self.browser.find_element(By.XPATH, "//input[@name='fear_level'][@value='4']").click()
        self.browser.find_element(By.XPATH, "//input[@name='sadness_level'][@value='6']").click()
        self.browser.find_element(By.XPATH, "//input[@name='joy_level'][@value='5']").click()
        self.browser.find_element(By.XPATH, "//input[@name='energy_level'][@value='7']").click()
        
        self.browser.find_element(By.NAME, "main_complaint").send_keys("Sente uma ansiedade generalizada e constante, com picos de pânico em situações de estresse no trabalho.")
        self.browser.find_element(By.NAME, "complaint_history").send_keys("Os sintomas se intensificaram nos últimos 8 meses, após uma mudança de responsabilidades no emprego.")
        self.browser.find_element(By.NAME, "signs_symptoms").send_keys("Aperto no peito, taquicardia, insônia, irritabilidade, dificuldade de concentração.")
        self.browser.find_element(By.NAME, "current_treatment").send_keys("Nenhum tratamento medicamentoso no momento. Tentou meditação através de aplicativos.")
        self.browser.find_element(By.NAME, "notes").send_keys("Paciente demonstra bom insight sobre os gatilhos de sua ansiedade, mas apresenta dificuldade em estabelecer limites.")
        self._take_screenshot("formulario_novo_paciente_preenchido")

        logger.info("Admin: Clicando no botão principal para salvar o paciente...")
        save_button = self.browser.find_element(By.XPATH, "//div[@class='form-actions']/button[@type='submit']")
        save_button.click()

        logger.info("Admin: A buscar o paciente recém-criado para obter o link...")
        self.wait.until(EC.url_contains('/admin/patients'))
        search_box = self.wait.until(EC.presence_of_element_located((By.ID, "search-box")))
        search_box.send_keys(novo_paciente_nome)
        search_box.submit()
        
        # Extrai o token do botão "Ver Link"
        ver_link_btn = self.wait.until(EC.presence_of_element_located((By.XPATH, f"//td[text()='{novo_paciente_nome}']/parent::tr//button[text()='Ver Link']")))
        onclick_attr = ver_link_btn.get_attribute("onclick")
        token_match = re.search(r"showConsentLink\('(.+?)'\)", onclick_attr)
        self.assertIsNotNone(token_match, "Não foi possível extrair o token do botão 'Ver Link'")
        token = token_match.group(1)
        consent_url = f"{BASE_URL}/portal/login/{token}"
        logger.info(f"Link de consentimento extraído: {consent_url}")

        # Abre o link em uma nova aba
        logger.info("Paciente: Abrindo o link de consentimento em uma nova aba...")
        self.browser.switch_to.new_window('tab')
        self.browser.get(consent_url)
        
        # Clica no botão "Acessar" na página de login do token
        logger.info("Paciente: Clicando no botão 'Acessar' para validar o token...")
        self.wait.until(EC.element_to_be_clickable((By.CSS_SELECTOR, "button.btn-submit"))).click()

        # Preenche e envia o formulário de consentimento
        logger.info("Paciente: Preenchendo o formulário de consentimento...")
        self.wait.until(EC.presence_of_element_located((By.ID, "consent_cpf_rg_inline"))).send_keys(TEST_CPF)
        self.browser.find_element(By.XPATH, "//input[@name='how_found'][@value='Google']").click()
        self._take_screenshot("formulario_consentimento")
        self.browser.find_element(By.CLASS_NAME, "btn-submit").click()

        # Verifica a página de sucesso
        self.wait.until(EC.url_contains("/portal/success"))
        self._take_screenshot("pagina_sucesso_consentimento")
        logger.info("Paciente: Formulário de consentimento enviado com sucesso.")

        # Fecha a aba do paciente e volta para a do admin
        self.browser.close()
        self.browser.switch_to.window(self.browser.window_handles[0])

        # Verifica se o status do consentimento foi atualizado
        logger.info("Admin: Verificando a atualização do status do consentimento...")
        self.browser.refresh()
        status_element = self.wait.until(EC.presence_of_element_located((By.XPATH, f"//td[text()='{novo_paciente_nome}']/parent::tr/td[5]")))
        self.assertIn("Fornecido", status_element.text)
        self._take_screenshot("status_consentimento_atualizado")
        logger.info("Admin: Status do consentimento verificado com sucesso.")
        
        # 4. Agendamento de Consulta e Verificação de Pagamento
        logger.info(f"Admin: Acedendo ao perfil do paciente '{novo_paciente_nome}' para agendar uma consulta...")
        linha_paciente = self.wait.until(EC.presence_of_element_located((By.XPATH, f"//td[text()='{novo_paciente_nome}']/parent::tr")))
        linha_paciente.find_element(By.LINK_TEXT, "Ver Perfil e Agenda").click()

        self.wait.until(EC.url_contains('/admin/patients/profile/'))
        self._take_screenshot("perfil_paciente_para_agendamento")

        logger.info("Admin: Preenchendo o formulário de agendamento...")
        
        appointment_date_sequence = "9222025"
        appointment_time = "13:00"
        
        Select(self.browser.find_element(By.ID, "doctor_id")).select_by_visible_text(TERAPEUTA_NOME)
        self.browser.find_element(By.ID, "appointment_date").send_keys(appointment_date_sequence)
        self.browser.find_element(By.ID, "start_time").send_keys(appointment_time)
        self.browser.find_element(By.ID, "price").send_keys("200.00")
        self.browser.find_element(By.ID, "notes").send_keys("Paciente agendado via teste automatizado. Sessão inicial.")
        
        self._take_screenshot("formulario_agendamento_preenchido")
        
        schedule_button = self.wait.until(EC.element_to_be_clickable((By.XPATH, "//fieldset[legend[text()='Agendar Nova Consulta']]//button[@type='submit']")))
        schedule_button.click()

        logger.info("Admin: Verificando se a consulta foi agendada com sucesso...")
        
        self.wait.until(EC.presence_of_element_located((By.XPATH, "//legend[text()='Consultas Agendadas']")))

        expected_datetime_str = f"22/09/2025 {appointment_time}"
        appointment_row = self.wait.until(EC.presence_of_element_located((By.XPATH, f"//td[contains(text(), '{expected_datetime_str}')]/parent::tr")))
        
        self.assertIn(TERAPEUTA_NOME, appointment_row.text)
        self.assertIn("200.00", appointment_row.text)
        self.assertIn("pendente", appointment_row.text)
        self._take_screenshot("consulta_agendada_com_sucesso")
        logger.info("Admin: Consulta verificada com sucesso na lista.")

        # 5. Marcar consulta como paga
        logger.info("Admin: Marcando a consulta como paga...")
        appointment_row.find_element(By.LINK_TEXT, "Marcar como Pago").click()

        # Espera a página recarregar e verifica o novo status
        logger.info("Admin: Verificando se o status do pagamento foi atualizado...")
        updated_appointment_row = self.wait.until(EC.presence_of_element_located((By.XPATH, f"//td[contains(text(), '{expected_datetime_str}')]/parent::tr")))
        
        payment_status_cell = updated_appointment_row.find_element(By.XPATH, "./td[4]") # A quarta célula (Pagamento)
        self.assertIn("✅", payment_status_cell.text)
        self.assertIn("pago", payment_status_cell.text)
        self._take_screenshot("consulta_marcada_como_paga")
        logger.info("Admin: Status de pagamento verificado com sucesso.")

        # 6. Gerar Resumo de IA
        logger.info(f"Admin: Acedendo ao prontuário do paciente '{novo_paciente_nome}' para gerar resumo de IA...")
        self.browser.get(f'{BASE_URL}/admin/patients')
        search_box_for_ai = self.wait.until(EC.presence_of_element_located((By.ID, "search-box")))
        search_box_for_ai.clear()
        search_box_for_ai.send_keys(novo_paciente_nome)
        search_box_for_ai.submit()

        linha_paciente_for_ai = self.wait.until(EC.presence_of_element_located((By.XPATH, f"//td[text()='{novo_paciente_nome}']/parent::tr")))
        linha_paciente_for_ai.find_element(By.LINK_TEXT, "Ver Perfil e Agenda").click()

        self.wait.until(EC.url_contains('/admin/patients/profile/'))
        edit_prontuario_link = self.wait.until(EC.element_to_be_clickable((By.LINK_TEXT, "Editar Dados / Ver Prontuário")))
        edit_prontuario_link.click()

        self.wait.until(EC.url_contains('/admin/patients/edit/'))
        logger.info("Admin: Clicando no botão para gerar resumo de IA...")
        ai_button = self.wait.until(EC.element_to_be_clickable((By.ID, "btn-ai-summary")))
        ai_button.click()

        logger.info("Admin: A aguardar a resposta da IA...")
        summary_container = self.wait.until(EC.presence_of_element_located((By.ID, "ai-summary-container")))
        
        # Usa um timeout maior para a IA
        WebDriverWait(self.browser, AI_TIMEOUT).until(
            lambda driver: "aguarde" not in summary_container.text and len(summary_container.text) > 20
        )

        summary_text = summary_container.text
        self.assertIn("Temas Recorrentes", summary_text)
        self.assertIn("Evolução", summary_text)
        self._take_screenshot("resumo_ia_gerado_pelo_admin")
        logger.info(f"Admin: Resumo de IA gerado com sucesso! Conteúdo: {summary_text[:100]}...")

        # Limpeza: Remove o paciente criado
        logger.info(f"Admin: Removendo paciente de teste '{novo_paciente_nome}'...")
        self.browser.get(f'{BASE_URL}/admin/patients')
        search_box_final = self.wait.until(EC.presence_of_element_located((By.ID, "search-box")))
        search_box_final.send_keys(novo_paciente_nome)
        search_box_final.submit()
        
        linha_paciente_final = self.wait.until(EC.presence_of_element_located((By.XPATH, f"//td[text()='{novo_paciente_nome}']/parent::tr")))
        linha_paciente_final.find_element(By.LINK_TEXT, "Remover").click()
        self.wait.until(EC.alert_is_present()).accept()
        self.wait.until(EC.staleness_of(linha_paciente_final))
        logger.info("Admin: Paciente de teste removido.")

        self._logout()

if __name__ == '__main__':
    unittest.main(verbosity=2, failfast=True)
