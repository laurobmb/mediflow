// Anima as barras de progresso ao carregar a página
document.addEventListener('DOMContentLoaded', function() {
    const bars = document.querySelectorAll('.bar');
    // Adiciona um pequeno atraso para garantir que a transição seja visível
    setTimeout(() => {
        bars.forEach(bar => {
            const width = bar.getAttribute('data-width');
            bar.style.width = width + '%';
        });
    }, 100);
});
