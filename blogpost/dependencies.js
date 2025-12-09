document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('dependency-form');
    const generateBtn = document.getElementById('generate-btn');
    const verboseToggle = document.getElementById('verbose-toggle');
    const resultsContainer = document.getElementById('results-container');
    const loadingSpinner = document.getElementById('loading-spinner');
    const diagramOutput = document.getElementById('diagram-output');
    const jsonOutput = document.getElementById('json-output');
    const errorMessage = document.getElementById('error-message');

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        resultsContainer.style.display = 'block';
        loadingSpinner.style.display = 'block';
        diagramOutput.style.display = 'none';
        jsonOutput.style.display = 'none';
        errorMessage.style.display = 'none';
        diagramOutput.innerHTML = '';
        jsonOutput.textContent = '';
        errorMessage.textContent = '';
        generateBtn.disabled = true;
        generateBtn.querySelector('span').textContent = 'Generating...';

        const identifier = document.getElementById('project-identifier').value;
        const sourceType = document.getElementById('source-type').value;

        try {
            const depsResponse = await fetch(`/api/dependencies?source=${sourceType}&id=${identifier}`);
            
            if (!depsResponse.ok) {
                const errorText = await depsResponse.text();
                throw new Error(errorText || 'Failed to fetch dependencies.');
            }
            
            const jsonData = await depsResponse.text();

            if (verboseToggle.checked) {
                const parsedJson = JSON.parse(jsonData);
                jsonOutput.textContent = JSON.stringify(parsedJson, null, 2);
                jsonOutput.style.display = 'block';
            }

            const renderResponse = await fetch('/api/render', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: jsonData
            });

            if (!renderResponse.ok) {
                const errorText = await renderResponse.text();
                throw new Error(errorText || 'Failed to render SVG.');
            }
            
            const svgData = await renderResponse.text();
            
            diagramOutput.innerHTML = svgData;
            diagramOutput.style.display = 'block';

        } catch (error) {
            errorMessage.textContent = `Error: ${error.message}`;
            errorMessage.style.display = 'block';
        } finally {
            loadingSpinner.style.display = 'none';
            generateBtn.disabled = false;
            generateBtn.querySelector('span').textContent = 'Generate';
        }
    });
});
