document.addEventListener('DOMContentLoaded', function() {
    const searchBox = document.getElementById('search-box');
    const resultsContainer = document.getElementById('autocomplete-results');

    // Verifica se os elementos existem na página antes de adicionar os listeners
    if (searchBox && resultsContainer) {
        searchBox.addEventListener('input', function() {
            const term = this.value;
            if (term.length < 2) {
                resultsContainer.innerHTML = '';
                resultsContainer.style.display = 'none';
                return;
            }

            fetch(`/admin/patients/search?term=${encodeURIComponent(term)}`)
                .then(response => response.json())
                .then(data => {
                    resultsContainer.innerHTML = '';
                    if (data && data.length > 0) {
                        resultsContainer.style.display = 'block';
                        data.forEach(name => {
                            const div = document.createElement('div');
                            div.textContent = name;
                            div.addEventListener('click', function() {
                                searchBox.value = this.textContent;
                                resultsContainer.innerHTML = '';
                                resultsContainer.style.display = 'none';
                                searchBox.form.submit();
                            });
                            resultsContainer.appendChild(div);
                        });
                    } else {
                        resultsContainer.style.display = 'none';
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    resultsContainer.style.display = 'none';
                });
        });

        // Esconde os resultados se o utilizador clicar fora da caixa de busca
        document.addEventListener('click', function(e) {
            if (e.target !== searchBox) {
                resultsContainer.innerHTML = '';
                resultsContainer.style.display = 'none';
            }
        });
    }
});
