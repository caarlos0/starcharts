(function () {
    new ClipboardJS('.copy-btn');
    hljs.highlightAll();
    document.querySelectorAll('time').forEach(t => timeago.render(t));
    document.querySelectorAll('input#repository').forEach(i => i.select());

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

    const chart = document.querySelector('#chart');
    const code = document.querySelector('code');
    document.querySelectorAll('.button-group button').forEach(group => {
        group.addEventListener('click', () => {
            const variant = group.dataset.variant;
            chart.src = `${chart.dataset.src}?variant=${variant}`;
        });
    });
})();