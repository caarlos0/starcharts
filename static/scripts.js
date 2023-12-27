(function () {
    const clipboard = new ClipboardJS('.copy-btn');
    let prevValue = null;
    let timeoutRef = null;

    clipboard.on('success', function (e) {
        e.clearSelection();

        if (timeoutRef) clearTimeout(timeoutRef);
        if (!prevValue) prevValue = e.trigger.innerText;

        e.trigger.innerText = 'Copied!';
        timeoutRef = setTimeout(function () {
            e.trigger.innerText = prevValue;
            prevValue = null;
            timeoutRef = null;
        }, 1000);
    });

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
            '#101010',
            '#6b63ff',
            '#e76060',
            '#2f81f7',
            '#333333',
        ],
    });

    document.querySelectorAll('time').forEach(function (t) {
        return timeago.render(t);
    });

    document.querySelectorAll('input#repository').forEach(function (i) {
        return i.select();
    });

    const colorInputElements = Array.from(document.querySelectorAll('[data-coloris]'));
    colorInputElements.forEach(function (input) {
        const prevValue = localStorage.getItem(input.name);
        if (prevValue) input.value = prevValue;
    });


    const codeElement = document.querySelector('.code-block code');
    const codeTemplateElement = document.querySelector('#code-template');
    const chartElement = document.querySelector('#chart');

    const chartUrl = chartElement.dataset.src;
    const codeTemplate = codeTemplateElement.innerText;

    function refreshState(variant) {
        const url = new URL(chartUrl, document.location.origin);
        if (variant === 'custom') {
            colorInputElements.forEach(function (color) {
                return url.searchParams.set(color.name, color.value);
            });
        } else {
            url.searchParams.set('variant', variant);
        }

        chartElement.src = url.toString();
        codeElement.innerHTML = codeTemplate.replace('$URL', url.toString());
        hljs.highlightAll();
    }

    colorInputElements.forEach(function (element) {
        element.addEventListener('change', function () {
            localStorage.setItem(element.name, element.value);
            refreshState('custom');
        });
    });

    const groupElements = document.querySelectorAll('.button-group');
    groupElements.forEach(function (group) {
        const buttonElements = group.querySelectorAll('button');
        buttonElements.forEach(function (button) {
            button.addEventListener('click', function () {
                buttonElements.forEach(function (b) {
                    b.classList.remove('active');
                });
                button.classList.add('active');
            });
        });
    });


    const customisationElement = document.querySelector('.customisation');
    document.querySelectorAll('.button-group button').forEach(function (group) {
        group.addEventListener('click', function () {
            if (group.dataset.variant === 'custom') {
                customisationElement.classList.add('opened');
                refreshState(group.dataset.variant);
            } else {
                customisationElement.classList.remove('opened');
                refreshState(group.dataset.variant);
            }
        });
    });

    refreshState('adaptive');
})();
