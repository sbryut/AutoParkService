document.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('.dropdown-toggle').forEach(item => {
        item.addEventListener('click', (e) => {
            e.preventDefault();
            const menu = e.target.nextElementSibling;
            menu.classList.toggle('show');
        });
    });
});

const deleteForms = document.querySelectorAll('form[action][method="POST"] input[value="DELETE"]');
deleteForms.forEach(input => {
    const form = input.closest('form');
    if (!form) return;

    form.addEventListener('submit', (e) => {
        e.preventDefault();

        const entityId = form.querySelector("input[name='id']").value;

        const fullUrl = form.action;
        const urlParts = new URL(fullUrl).pathname.split('/').filter(part => part);
        const entityType = urlParts[0]; // первый сегмент пути

        const confirmMessage = `Вы уверены, что хотите удалить ${getEntityName(entityType)}?`;

        if (confirm(confirmMessage)) {
            const formData = new FormData(form);

            const deleteUrl = `/${entityType}/${entityId}/delete`;

            fetch(deleteUrl, {
                method: 'POST',
                body: formData,
            }).then(response => {
                if (response.ok) {
                    const tableRow = form.closest('tr');
                    if (tableRow) {
                        tableRow.remove();
                    } else {
                        window.location.reload();
                    }
                } else {
                    return response.text().then(errorText => {
                        throw new Error(errorText || 'Ошибка при удалении');
                    });
                }
            }).catch(error => {
                console.error('Ошибка:', error);
                alert(error.message || 'Ошибка при удалении');
            });
        }
    });
});

function getEntityName(entityType) {
    switch(entityType) {
        case 'drivers': return 'водителя';
        case 'autos': return 'автомобиль';
        case 'routes': return 'маршрут';
        default: return 'эту запись';
    }
}

