// static/js/ai_summary.js

document.addEventListener('DOMContentLoaded', function() {
    const controls = document.getElementById('ai-summary-controls');
    if (!controls) return;

    const btn = document.getElementById('btn-ai-summary');
    const container = document.getElementById('ai-summary-container');

    // Lê os dados diretamente do HTML usando os atributos data-*
    const userType = controls.dataset.usertype;
    const patientId = controls.dataset.patientid;

    if (btn && container && patientId && userType) {
        btn.addEventListener('click', function() {
            // Decide a URL correta com base no tipo de usuário lido do HTML
            let apiUrl = '';
            if (userType === 'admin') {
                apiUrl = `/admin/pacientes/${patientId}/ai-summary`;
            } else if (userType === 'terapeuta') {
                apiUrl = `/terapeuta/pacientes/${patientId}/ai-summary`;
            } else {
                console.error("Tipo de usuário desconhecido para gerar resumo de IA:", userType);
                container.innerHTML = `<p style="color: red;">Erro: Perfil de usuário desconhecido.</p>`;
                return;
            }

            // Mostra o status de carregamento
            btn.disabled = true;
            btn.textContent = 'Analisando prontuário...';
            container.style.display = 'block';
            container.innerHTML = '<p>Por favor, aguarde enquanto a IA gera o resumo...</p>';

            // Faz a chamada para a API
            fetch(apiUrl)
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Falha na resposta da rede.');
                    }
                    return response.json();
                })
                .then(data => {
                    if (data.error) {
                         container.innerHTML = `<p style="color: red;"><strong>Erro:</strong> ${data.error}</p>`;
                    } else {
                        // Converte o markdown básico da IA para HTML
                        let htmlSummary = data.summary.toString()
                            .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>') // Negrito
                            .replace(/\*/g, '') // Remove asteriscos de bullets
                            .replace(/\n/g, '<br>'); // Novas linhas
                        container.innerHTML = htmlSummary;
                    }
                })
                .catch(error => {
                    console.error('Erro ao buscar resumo da IA:', error);
                    container.innerHTML = `<p style="color: red;">Ocorreu um erro na comunicação com o serviço de IA.</p>`;
                })
                .finally(() => {
                    // Esconde o botão após o uso
                    btn.style.display = 'none';
                });
        });
    }
});