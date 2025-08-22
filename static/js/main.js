// Função para mostrar/esconder campos baseando-se na seleção de radio buttons
function toggleField(fieldId, show) {
    const field = document.getElementById(fieldId);
    if (field) {
        // Ajusta o display baseado no tipo de elemento se precisar de flexbox ou block
        field.style.display = show ? 'block' : 'none';
        if (!show) {
            field.value = ''; // Limpa o valor quando escondido
        }
    }
}

// Função específica para "Filho(s): Quantos?"
function toggleChildrenQty(show) {
    const qtyGroup = document.getElementById('children_qty_group');
    const numChildrenInput = document.getElementById('num_children');
    if (qtyGroup) {
        qtyGroup.style.display = show ? 'flex' : 'none'; // Use flex para manter o alinhamento em form-row
        if (!show) {
            numChildrenInput.value = '';
        }
    }
}

// Função específica para "Transtorno mental: Faz tratamento?"
function toggleMentalDisorderTreatment(show) {
    const treatmentGroup = document.getElementById('mental_disorder_treatment_group');
    if (treatmentGroup) {
        treatmentGroup.style.display = show ? 'block' : 'none';
        // Reseta campos internos se o grupo pai for escondido
        if (!show) {
            const treatmentRadios = document.querySelectorAll('input[name="mental_disorder_treatment"]');
            treatmentRadios.forEach(radio => radio.checked = false);
            toggleField('mental_disorder_details', false);
        }
    }
}

// Inicializa o estado dos campos dinâmicos e o cálculo de idade ao carregar a página
document.addEventListener('DOMContentLoaded', (event) => {

    // --- INICIALIZAÇÃO ROBUSTA DE CAMPOS DINÂMICOS ---
    // Em vez de um bloco de código para cada campo, usamos uma configuração.
    // Para adicionar um novo campo, basta adicionar um novo objeto a esta lista.
    const dynamicFieldsConfig = [
        { radioName: 'children', triggerValue: 'Sim', targetId: 'children_qty_group', toggleFunction: toggleChildrenQty },
        { radioName: 'repetitive_effort', triggerValue: 'Sim', targetId: 'repetitive_effort_spec', toggleFunction: toggleField },
        { radioName: 'physical_activity', triggerValue: 'Sim', targetId: 'physical_activity_spec', toggleFunction: toggleField },
        { radioName: 'alcohol', triggerValue: 'Sim', targetId: 'alcohol_frequency', toggleFunction: toggleField },
        { radioName: 'mental_disorder', triggerValue: 'Sim', targetId: 'mental_disorder_treatment_group', toggleFunction: toggleMentalDisorderTreatment },
        { radioName: 'mental_disorder_treatment', triggerValue: 'Sim', targetId: 'mental_disorder_details', toggleFunction: toggleField },
        { radioName: 'religion', triggerValue: 'Sim', targetId: 'religion_details', toggleFunction: toggleField },
        { radioName: 'medication', triggerValue: 'Sim', targetId: 'medication_details', toggleFunction: toggleField },
        { radioName: 'surgery', triggerValue: 'Sim', targetId: 'surgery_details', toggleFunction: toggleField },
        { radioName: 'allergies', triggerValue: 'Sim', targetId: 'allergies_details', toggleFunction: toggleField },
        { radioName: 'how_found', triggerValue: 'Indicação de conhecido', targetId: 'referral_name', toggleFunction: toggleField },
        { radioName: 'how_found', triggerValue: 'Outro', targetId: 'other_source', toggleFunction: toggleField },
    ];

    dynamicFieldsConfig.forEach(config => {
        const radios = document.querySelectorAll(`input[name="${config.radioName}"]`);
        let initiallyShown = false;

        radios.forEach(radio => {
            // Verifica o estado inicial ao carregar a página
            if (radio.checked && radio.value === config.triggerValue) {
                initiallyShown = true;
            }
            // Adiciona o evento de mudança
            radio.addEventListener('change', () => {
                const shouldShow = document.querySelector(`input[name="${config.radioName}"]:checked`).value === config.triggerValue;
                config.toggleFunction(config.targetId, shouldShow);
            });
        });

        // Define o estado inicial do campo alvo
        config.toggleFunction(config.targetId, initiallyShown);
    });

    // Lógica especial para o campo "Fumante", que tem múltiplos gatilhos
    const smokerRadios = document.querySelectorAll('input[name="smoker"]');
    const smokerTimeField = document.getElementById('smoker_time');
    if (smokerRadios.length > 0 && smokerTimeField) {
        const updateSmokerField = () => {
            const selected = document.querySelector('input[name="smoker"]:checked');
            const show = selected && selected.value !== 'Não';
            toggleField('smoker_time', show);
        };
        smokerRadios.forEach(radio => radio.addEventListener('change', updateSmokerField));
        updateSmokerField(); // Verifica o estado inicial
    }


    // --- FUNCIONALIDADE: CÁLCULO AUTOMÁTICO DE IDADE ---
    const dobInput = document.getElementById('dob');
    const ageInput = document.getElementById('age');

    if (dobInput && ageInput) {
        // Função para calcular a idade
        const calculateAge = () => {
            const dobValue = dobInput.value;
            if (dobValue) {
                const birthDate = new Date(dobValue);
                const today = new Date();
                let age = today.getFullYear() - birthDate.getFullYear();
                const m = today.getMonth() - birthDate.getMonth();
                if (m < 0 || (m === 0 && today.getDate() < birthDate.getDate())) {
                    age--;
                }
                ageInput.value = age >= 0 ? age : '';
            } else {
                ageInput.value = '';
            }
        };

        // Adiciona o evento para calcular a idade sempre que a data for alterada
        dobInput.addEventListener('change', calculateAge);

        // Calcula a idade ao carregar a página, caso a data já esteja preenchida (no modo de edição)
        calculateAge();
    }
});
