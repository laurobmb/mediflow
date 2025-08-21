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

// Inicializa o estado dos campos dinâmicos ao carregar a página
document.addEventListener('DOMContentLoaded', (event) => {
    // Filhos
    const childrenYes = document.querySelector('input[name="children"][value="Sim"]');
    if (!childrenYes || !childrenYes.checked) {
        toggleChildrenQty(false);
    }

    // Esforço Repetitivo
    const repetitiveEffortYes = document.querySelector('input[name="repetitive_effort"][value="Sim"]');
    if (!repetitiveEffortYes || !repetitiveEffortYes.checked) {
        toggleField('repetitive_effort_spec', false);
    }

    // Atividade Física
    const physicalActivityYes = document.querySelector('input[name="physical_activity"][value="Sim"]');
    if (!physicalActivityYes || !physicalActivityYes.checked) {
        toggleField('physical_activity_spec', false);
    }

    // Fumante
    const smokerRadios = document.querySelectorAll('input[name="smoker"]');
    let smokerChecked = false;
    smokerRadios.forEach(radio => { if (radio.checked && radio.value !== 'Não') smokerChecked = true; });
    if (!smokerChecked) {
        toggleField('smoker_time', false);
    }

    // Bebidas Alcoólicas
    const alcoholYes = document.querySelector('input[name="alcohol"][value="Sim"]');
    if (!alcoholYes || !alcoholYes.checked) {
        toggleField('alcohol_frequency', false);
    }

    // Transtorno mental
    const mentalDisorderYes = document.querySelector('input[name="mental_disorder"][value="Sim"]');
    if (!mentalDisorderYes || !mentalDisorderYes.checked) {
        toggleMentalDisorderTreatment(false);
    } else { // Se "Sim" estiver marcado, verifica o tratamento
        const treatmentYes = document.querySelector('input[name="mental_disorder_treatment"][value="Sim"]');
        if (!treatmentYes || !treatmentYes.checked) {
            toggleField('mental_disorder_details', false);
        }
    }

    // Religião
    const religionYes = document.querySelector('input[name="religion"][value="Sim"]');
    if (!religionYes || !religionYes.checked) {
        toggleField('religion_details', false);
    }

    // Medicamentos
    const medicationYes = document.querySelector('input[name="medication"][value="Sim"]');
    if (!medicationYes || !medicationYes.checked) {
        toggleField('medication_details', false);
    }

    // Cirurgias
    const surgeryYes = document.querySelector('input[name="surgery"][value="Sim"]');
    if (!surgeryYes || !surgeryYes.checked) {
        toggleField('surgery_details', false);
    }

    // Alergias
    const allergiesYes = document.querySelector('input[name="allergies"][value="Sim"]');
    if (!allergiesYes || !allergiesYes.checked) {
        toggleField('allergies_details', false);
    }

    // Como Encontrou (indicação e outro)
    const howFoundReferral = document.querySelector('input[name="how_found"][value="Indicação de conhecido"]');
    if (!howFoundReferral || !howFoundReferral.checked) {
        toggleField('referral_name', false);
    }
    const howFoundOther = document.querySelector('input[name="how_found"][value="Outro"]');
    if (!howFoundOther || !howFoundOther.checked) {
        toggleField('other_source', false);
    }
});