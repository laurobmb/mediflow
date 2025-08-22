document.addEventListener('DOMContentLoaded', function() {
    const searchBox = document.getElementById('search-box');
    const resultsContainer = document.getElementById('autocomplete-results');

    if (searchBox && resultsContainer) {
        // Pega a URL da API a partir de um atributo no próprio input
        const apiUrl = searchBox.dataset.apiUrl;
        if (!apiUrl) {
            console.error('O campo de busca não possui o atributo data-api-url.');
            return;
        }

        searchBox.addEventListener('input', function() {
            const term = this.value;
            if (term.length < 2) {
                resultsContainer.innerHTML = '';
                resultsContainer.style.display = 'none';
                return;
            }

            // Usa a URL dinâmica que foi lida do atributo HTML
            fetch(`${apiUrl}?term=${encodeURIComponent(term)}`)
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
                                resultsContainer.style.display = 'none';
                                searchBox.form.submit();
                            });
                            resultsContainer.appendChild(div);
                        });
                    } else {
                        resultsContainer.style.display = 'none';
                    }
                });
        });

        document.addEventListener('click', function(e) {
            if (e.target !== searchBox) {
                resultsContainer.style.display = 'none';
            }
        });
    }
});