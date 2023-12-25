(function () {
    new ClipboardJS('.copy-btn');
    hljs.highlightAll();

    Coloris({
        theme: 'pill',
        themeMode: 'auto',
        alpha: true,
        margin: 16,
        format: 'hex',
        formatToggle: false,
        closeButton: true,
        closeLabel: 'Apply',
        swatches: [
            '#FFFFFF',
            '#333333',
            '#6b63ff',
        ],
    });

    document.querySelectorAll('time').forEach(t => timeago.render(t));
    document.querySelectorAll('input#repository').forEach(i => i.select());

    const chart = document.querySelector('#chart');
    const url = new URL(chart.dataset.src, document.location.origin);
    const customisation = document.querySelector('.customisation');

    const chartColorInputs = Array.from(document.querySelectorAll('[data-coloris]'));
    const refreshChart = () => {
        url.searchParams.size = 0;
        chartColorInputs.forEach(color => url.searchParams.set(color.name, color.value))
        chart.src = url.toString();
    }
    chartColorInputs.forEach(element => {
        element.addEventListener('change', () => {
            refreshChart();
        });
    });

    const groups = document.querySelectorAll('.button-group');
    groups.forEach(group => {
        const buttons = group.querySelectorAll('button');
        buttons.forEach(button => {
            button.addEventListener('click', () => {
                buttons.forEach(b => b.classList.remove('active'));
                button.classList.add('active');
            });
        });
    });


    const code = document.querySelector('code');
    document.querySelectorAll('.button-group button').forEach(group => {
        group.addEventListener('click', () => {
            if (group.dataset.variant === 'custom') {
                customisation.classList.add('opened');
                refreshChart();
            } else {
                customisation.classList.remove('opened');
                const variant = group.dataset.variant;
                chart.src = `${chart.dataset.src}?variant=${variant}`;
            }
        });
    });
})();