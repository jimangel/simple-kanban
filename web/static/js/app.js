// Kanban Board Application
class KanbanApp {
    constructor() {
        this.currentBoard = null;
        this.boards = [];
        this.lists = [];
        this.sortables = [];
        this.currentCard = null;
        this.apiBase = '/api';

        this.init();
    }

    async init() {
        this.setupEventListeners();
        this.currentBoard = 1; // Use board ID 1
        await this.loadBoardData();
    }

    setupEventListeners() {
        // Board title editing
        const boardTitle = document.getElementById('boardTitle');
        let originalTitle = boardTitle.textContent;

        boardTitle.addEventListener('focus', () => {
            originalTitle = boardTitle.textContent;
        });

        boardTitle.addEventListener('blur', async () => {
            const newTitle = boardTitle.textContent.trim();
            if (newTitle && newTitle !== originalTitle) {
                await this.updateBoardName(newTitle);
            } else if (!newTitle) {
                boardTitle.textContent = originalTitle;
            }
        });

        boardTitle.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                boardTitle.blur();
            } else if (e.key === 'Escape') {
                boardTitle.textContent = originalTitle;
                boardTitle.blur();
            }
        });

        // List management
        document.getElementById('addListBtn').addEventListener('click', () => this.showCreateListModal());
        document.getElementById('saveListBtn').addEventListener('click', () => this.createList());

        // Card management
        document.getElementById('saveCardBtn').addEventListener('click', () => this.saveCard());
        document.getElementById('deleteCardBtn').addEventListener('click', () => this.deleteCard());
        document.getElementById('archiveCardBtn').addEventListener('click', () => this.archiveCard());
        document.getElementById('addCommentBtn').addEventListener('click', () => this.addComment());

        // Search and archive
        document.getElementById('searchBtn').addEventListener('click', () => this.showSearchModal());
        document.getElementById('searchInput').addEventListener('input', (e) => this.searchCards(e.target.value));
        document.getElementById('archiveBtn').addEventListener('click', () => this.showArchiveModal());

        // Enter key for comments
        document.getElementById('newComment').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.addComment();
            }
        });
    }

    // API Methods
    async apiCall(url, method = 'GET', body = null) {
        const options = {
            method,
            headers: {
                'Content-Type': 'application/json',
            },
        };

        if (body) {
            options.body = JSON.stringify(body);
        }

        try {
            const response = await fetch(this.apiBase + url, options);
            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'API call failed');
            }
            const data = await response.json();
            // Ensure null is converted to empty array for array responses
            return data === null ? [] : data;
        } catch (error) {
            console.error('API Error:', error);
            this.showAlert('Error: ' + error.message, 'danger');
            throw error;
        }
    }

    // Board Management
    async loadBoardData() {
        try {
            const board = await this.apiCall(`/boards/${this.currentBoard}`);
            document.getElementById('boardTitle').textContent = board.name;

            const lists = await this.apiCall(`/boards/${this.currentBoard}/lists`);
            this.lists = lists;
            await this.renderBoard();
        } catch (error) {
            console.error('Failed to load board:', error);
            this.showAlert('Failed to load board. Using default board name.', 'warning');
            // Even if board load fails, try to load lists
            try {
                const lists = await this.apiCall(`/boards/${this.currentBoard}/lists`);
                this.lists = lists;
                await this.renderBoard();
            } catch (listError) {
                console.error('Failed to load lists:', listError);
            }
        }
    }

    async updateBoardName(name) {
        try {
            await this.apiCall(`/boards/${this.currentBoard}`, 'PUT', { name });
            this.showAlert('Board name updated', 'success');
        } catch (error) {
            console.error('Failed to update board name:', error);
            this.showAlert('Failed to update board name', 'danger');
        }
    }

    async loadBoards() {
        try {
            this.boards = await this.apiCall('/boards');
            this.renderBoardSelector();

            if (this.boards.length > 0) {
                await this.selectBoard(this.boards[0].id);
            }
        } catch (error) {
            console.error('Failed to load boards:', error);
        }
    }

    renderBoardSelector() {
        const selector = document.getElementById('boardSelector');
        selector.innerHTML = '';

        if (this.boards.length === 0) {
            selector.innerHTML = '<option value="">No boards available</option>';
            return;
        }

        this.boards.forEach(board => {
            const option = document.createElement('option');
            option.value = board.id;
            option.textContent = board.name;
            if (board.description) {
                option.title = board.description;
            }
            selector.appendChild(option);
        });
    }

    async selectBoard(boardId) {
        if (!boardId) return;

        try {
            this.currentBoard = parseInt(boardId);
            const lists = await this.apiCall(`/boards/${boardId}/lists`);
            this.lists = lists;
            await this.renderBoard();
        } catch (error) {
            console.error('Failed to select board:', error);
        }
    }

    showCreateBoardModal() {
        const modal = new bootstrap.Modal(document.getElementById('createBoardModal'));
        document.getElementById('boardName').value = '';
        document.getElementById('boardDescription').value = '';
        modal.show();
    }

    async createBoard() {
        const name = document.getElementById('boardName').value.trim();
        const description = document.getElementById('boardDescription').value.trim();

        if (!name) {
            this.showAlert('Board name is required', 'warning');
            return;
        }

        try {
            const board = await this.apiCall('/boards', 'POST', { name, description });
            this.boards.push(board);
            this.renderBoardSelector();
            await this.selectBoard(board.id);

            const modal = bootstrap.Modal.getInstance(document.getElementById('createBoardModal'));
            modal.hide();

            this.showAlert('Board created successfully', 'success');
        } catch (error) {
            console.error('Failed to create board:', error);
        }
    }

    // List Management
    async renderBoard() {
        const boardContainer = document.getElementById('kanbanBoard');
        boardContainer.innerHTML = '';

        // Clear old sortables
        this.sortables.forEach(sortable => sortable.destroy());
        this.sortables = [];

        if (this.lists.length === 0) {
            boardContainer.innerHTML = `
                <div class="empty-state">
                    <h3>No lists yet</h3>
                    <p>Click "Add List" to create your first list</p>
                </div>
            `;
            return;
        }

        for (const list of this.lists) {
            const listElement = this.createListElement(list);
            boardContainer.appendChild(listElement);
            // Load cards AFTER the element is in the DOM
            await this.loadCards(list.id);
        }

        // Make the board itself sortable for lists
        Sortable.create(boardContainer, {
            animation: 150,
            handle: '.kanban-list-header',
            onEnd: (evt) => this.moveList(evt)
        });
    }

    createListElement(list) {
        const listDiv = document.createElement('div');
        listDiv.className = 'kanban-list';
        listDiv.dataset.listId = list.id;
        listDiv.style.borderTop = `4px solid ${list.color || '#6b7280'}`;

        listDiv.innerHTML = `
            <div class="kanban-list-header">
                <h5 class="kanban-list-title" contenteditable="true" spellcheck="false" data-list-id="${list.id}">${this.escapeHtml(list.name)}</h5>
                <div class="kanban-list-actions">
                    <button onclick="app.deleteList(${list.id})" title="Delete List">
                        <i class="bi bi-trash"></i>
                    </button>
                </div>
            </div>
            <div class="kanban-cards" data-list-id="${list.id}">
                <!-- Cards will be loaded here -->
            </div>
            <button class="add-card-btn" onclick="app.showQuickAddCard(${list.id})">
                <i class="bi bi-plus"></i> Add Card
            </button>
        `;

        // Add event listeners for inline editing of list title
        const listTitle = listDiv.querySelector('.kanban-list-title');
        let originalListName = list.name;

        listTitle.addEventListener('focus', () => {
            originalListName = listTitle.textContent;
        });

        listTitle.addEventListener('blur', async () => {
            const newName = listTitle.textContent.trim();
            if (newName && newName !== originalListName) {
                await this.updateListName(list.id, newName);
            } else if (!newName) {
                listTitle.textContent = originalListName;
            }
        });

        listTitle.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                listTitle.blur();
            } else if (e.key === 'Escape') {
                listTitle.textContent = originalListName;
                listTitle.blur();
            }
        });

        // Make cards sortable
        const cardsContainer = listDiv.querySelector('.kanban-cards');
        const sortable = Sortable.create(cardsContainer, {
            group: 'cards',
            animation: 150,
            ghostClass: 'sortable-ghost',
            chosenClass: 'sortable-chosen',
            onEnd: (evt) => this.moveCard(evt)
        });
        this.sortables.push(sortable);

        return listDiv;
    }

    showCreateListModal() {
        if (!this.currentBoard) {
            this.showAlert('Please select a board first', 'warning');
            return;
        }

        const modal = new bootstrap.Modal(document.getElementById('createListModal'));
        document.getElementById('listName').value = '';
        document.getElementById('listColor').value = '#6b7280';
        modal.show();
    }

    async createList() {
        const name = document.getElementById('listName').value.trim();
        const color = document.getElementById('listColor').value;

        if (!name) {
            this.showAlert('List name is required', 'warning');
            return;
        }

        try {
            const list = await this.apiCall(`/boards/${this.currentBoard}/lists`, 'POST', { name, color });
            this.lists.push(list);
            await this.renderBoard();

            const modal = bootstrap.Modal.getInstance(document.getElementById('createListModal'));
            modal.hide();

            this.showAlert('List created successfully', 'success');
        } catch (error) {
            console.error('Failed to create list:', error);
        }
    }

    async deleteList(listId) {
        if (!confirm('Are you sure you want to delete this list and all its cards?')) {
            return;
        }

        try {
            await this.apiCall(`/lists/${listId}`, 'DELETE');
            this.lists = this.lists.filter(l => l.id !== listId);
            await this.renderBoard();
            this.showAlert('List deleted successfully', 'success');
        } catch (error) {
            console.error('Failed to delete list:', error);
        }
    }

    async updateListName(listId, name) {
        try {
            await this.apiCall(`/lists/${listId}`, 'PUT', { name });
            // Update local list data
            const list = this.lists.find(l => l.id === listId);
            if (list) {
                list.name = name;
            }
            this.showAlert('List name updated', 'success');
        } catch (error) {
            console.error('Failed to update list name:', error);
            this.showAlert('Failed to update list name', 'danger');
        }
    }

    async moveList(evt) {
        const listElement = evt.item;
        const listId = parseInt(listElement.dataset.listId);
        const newIndex = evt.newIndex;

        // Calculate new position
        const prevList = this.lists[newIndex - 1];
        const nextList = this.lists[newIndex + 1];
        let position;

        if (!prevList) {
            position = nextList ? nextList.position / 2 : 1;
        } else if (!nextList) {
            position = prevList.position + 1;
        } else {
            position = (prevList.position + nextList.position) / 2;
        }

        try {
            await this.apiCall(`/lists/${listId}/move`, 'PATCH', { position });
            // Reload lists to get updated positions
            const lists = await this.apiCall(`/boards/${this.currentBoard}/lists`);
            this.lists = lists;
        } catch (error) {
            console.error('Failed to move list:', error);
            await this.renderBoard(); // Revert on error
        }
    }

    // Card Management
    async loadCards(listId) {
        try {
            const cards = await this.apiCall(`/lists/${listId}/cards`);
            // Use more specific selector to target the cards container, not the outer list div
            const container = document.querySelector(`.kanban-cards[data-list-id="${listId}"]`);
            if (container) {
                container.innerHTML = '';
                // Add null-safety check for cards array
                if (cards && Array.isArray(cards)) {
                    cards.forEach(card => {
                        container.appendChild(this.createCardElement(card));
                    });
                }
            }
        } catch (error) {
            console.error('Failed to load cards:', error);
        }
    }

    createCardElement(card) {
        const cardDiv = document.createElement('div');
        cardDiv.className = 'kanban-card';
        cardDiv.dataset.cardId = card.id;
        cardDiv.dataset.listId = card.list_id;
        cardDiv.dataset.position = card.position;

        if (card.color) {
            cardDiv.style.borderLeft = `4px solid ${card.color}`;
        }

        let dueDateHtml = '';
        if (card.due_date) {
            const dueDate = new Date(card.due_date);
            const isOverdue = dueDate < new Date();
            dueDateHtml = `
                <div class="kanban-card-due ${isOverdue ? 'overdue' : ''}">
                    <i class="bi bi-calendar"></i>
                    ${dueDate.toLocaleDateString()}
                </div>
            `;
        }

        cardDiv.innerHTML = `
            <div class="kanban-card-title">${this.escapeHtml(card.title)}</div>
            ${card.description ? `<div class="kanban-card-description">${this.escapeHtml(card.description)}</div>` : ''}
            <div class="kanban-card-footer">
                ${dueDateHtml}
                <div class="kanban-card-actions">
                    <button class="btn btn-sm btn-link p-0" onclick="app.editCard(${card.id})">
                        <i class="bi bi-pencil"></i>
                    </button>
                </div>
            </div>
        `;

        return cardDiv;
    }

    showQuickAddCard(listId) {
        const listElement = document.querySelector(`.kanban-list[data-list-id="${listId}"]`);
        const addButton = listElement.querySelector('.add-card-btn');

        const quickAddDiv = document.createElement('div');
        quickAddDiv.className = 'quick-add-card';
        quickAddDiv.innerHTML = `
            <input type="text" placeholder="Enter card title..." id="quickCardTitle-${listId}" autofocus>
            <div class="quick-add-card-actions">
                <button class="btn btn-sm btn-primary" onclick="app.quickCreateCard(${listId})">Add</button>
                <button class="btn btn-sm btn-secondary" onclick="app.cancelQuickAdd(${listId})">Cancel</button>
            </div>
        `;

        addButton.style.display = 'none';
        addButton.parentNode.insertBefore(quickAddDiv, addButton);

        const input = document.getElementById(`quickCardTitle-${listId}`);
        input.focus();
        input.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.quickCreateCard(listId);
            } else if (e.key === 'Escape') {
                this.cancelQuickAdd(listId);
            }
        });
    }

    async quickCreateCard(listId) {
        const input = document.getElementById(`quickCardTitle-${listId}`);
        const title = input.value.trim();

        if (!title) {
            return;
        }

        try {
            const card = await this.apiCall(`/lists/${listId}/cards`, 'POST', { title });
            await this.loadCards(listId);
            this.cancelQuickAdd(listId);
            this.showAlert('Card created successfully', 'success');
        } catch (error) {
            console.error('Failed to create card:', error);
        }
    }

    cancelQuickAdd(listId) {
        const listElement = document.querySelector(`.kanban-list[data-list-id="${listId}"]`);
        const quickAddDiv = listElement.querySelector('.quick-add-card');
        const addButton = listElement.querySelector('.add-card-btn');

        if (quickAddDiv) {
            quickAddDiv.remove();
        }
        if (addButton) {
            addButton.style.display = 'block';
        }
    }

    async editCard(cardId) {
        try {
            const card = await this.apiCall(`/cards/${cardId}`);
            this.currentCard = card;

            const modal = bootstrap.Modal.getOrCreateInstance(document.getElementById('cardModal'));
            document.getElementById('cardModalTitle').textContent = 'Edit Card';
            document.getElementById('cardTitle').value = card.title || '';
            document.getElementById('cardDescription').value = card.description || '';
            document.getElementById('cardColor').value = card.color || '#ffffff';
            document.getElementById('cardDueDate').value = card.due_date ?
                new Date(card.due_date).toISOString().slice(0, 16) : '';

            document.getElementById('deleteCardBtn').style.display = 'inline-block';
            document.getElementById('archiveCardBtn').style.display = 'inline-block';
            document.getElementById('commentsSection').style.display = 'block';

            // Load comments
            await this.loadComments(cardId);

            modal.show();
        } catch (error) {
            console.error('Failed to load card:', error);
        }
    }

    async saveCard() {
        const title = document.getElementById('cardTitle').value.trim();
        const description = document.getElementById('cardDescription').value.trim();
        const color = document.getElementById('cardColor').value;
        const dueDate = document.getElementById('cardDueDate').value;

        if (!title) {
            this.showAlert('Card title is required', 'warning');
            return;
        }

        const cardData = {
            title,
            description,
            color
        };

        if (dueDate) {
            cardData.due_date = new Date(dueDate).toISOString();
        }

        try {
            if (this.currentCard) {
                // Update existing card
                await this.apiCall(`/cards/${this.currentCard.id}`, 'PUT', cardData);
                await this.loadCards(this.currentCard.list_id);
                this.showAlert('Card updated successfully', 'success');
            }

            const modal = bootstrap.Modal.getInstance(document.getElementById('cardModal'));
            modal.hide();
        } catch (error) {
            console.error('Failed to save card:', error);
        }
    }

    async deleteCard() {
        if (!this.currentCard) return;

        if (!confirm('Are you sure you want to delete this card?')) {
            return;
        }

        try {
            await this.apiCall(`/cards/${this.currentCard.id}`, 'DELETE');
            await this.loadCards(this.currentCard.list_id);

            const modal = bootstrap.Modal.getInstance(document.getElementById('cardModal'));
            modal.hide();

            this.showAlert('Card deleted successfully', 'success');
        } catch (error) {
            console.error('Failed to delete card:', error);
        }
    }

    async archiveCard() {
        if (!this.currentCard) return;

        try {
            await this.apiCall(`/cards/${this.currentCard.id}/archive`, 'POST');
            await this.loadCards(this.currentCard.list_id);

            const modal = bootstrap.Modal.getInstance(document.getElementById('cardModal'));
            modal.hide();

            this.showAlert('Card archived successfully', 'success');
        } catch (error) {
            console.error('Failed to archive card:', error);
        }
    }

    async moveCard(evt) {
        const cardElement = evt.item;
        const cardId = parseInt(cardElement.dataset.cardId);
        const newListId = parseInt(evt.to.dataset.listId);
        const newIndex = evt.newIndex;

        // Get all cards in the new list
        const cards = Array.from(evt.to.children);
        const prevCard = cards[newIndex - 1];
        const nextCard = cards[newIndex + 1];

        let position = 1;

        if (prevCard && nextCard) {
            // Calculate position between two cards
            position = (parseFloat(prevCard.dataset.position || 1) +
                       parseFloat(nextCard.dataset.position || 2)) / 2;
        } else if (prevCard) {
            position = parseFloat(prevCard.dataset.position || 1) + 1;
        } else if (nextCard) {
            position = parseFloat(nextCard.dataset.position || 1) / 2;
        }

        try {
            await this.apiCall(`/cards/${cardId}/move`, 'PATCH', {
                list_id: newListId,
                position: position
            });

            // Update card's data attributes
            cardElement.dataset.listId = newListId;
            cardElement.dataset.position = position;
        } catch (error) {
            console.error('Failed to move card:', error);
            // Revert the move on error
            evt.from.insertBefore(cardElement, evt.from.children[evt.oldIndex]);
        }
    }

    // Comments
    async loadComments(cardId) {
        try {
            const comments = await this.apiCall(`/cards/${cardId}/comments`);
            const container = document.getElementById('commentsList');
            container.innerHTML = '';

            comments.forEach(comment => {
                const commentDiv = document.createElement('div');
                commentDiv.className = 'comment';
                commentDiv.innerHTML = `
                    <div class="comment-content">${this.escapeHtml(comment.content)}</div>
                    <div class="comment-date">${new Date(comment.created_at).toLocaleString()}</div>
                `;
                container.appendChild(commentDiv);
            });
        } catch (error) {
            console.error('Failed to load comments:', error);
        }
    }

    async addComment() {
        if (!this.currentCard) return;

        const input = document.getElementById('newComment');
        const content = input.value.trim();

        if (!content) return;

        try {
            await this.apiCall(`/cards/${this.currentCard.id}/comments`, 'POST', { content });
            input.value = '';
            await this.loadComments(this.currentCard.id);
            this.showAlert('Comment added', 'success');
        } catch (error) {
            console.error('Failed to add comment:', error);
        }
    }

    // Search
    showSearchModal() {
        const modal = new bootstrap.Modal(document.getElementById('searchModal'));
        document.getElementById('searchInput').value = '';
        document.getElementById('searchResults').innerHTML = '';
        modal.show();
    }

    async searchCards(query) {
        if (!query) {
            document.getElementById('searchResults').innerHTML = '';
            return;
        }

        const includeArchived = document.getElementById('includeArchived').checked;

        try {
            const params = new URLSearchParams({
                query: query,
                board_id: this.currentBoard,
                archived: includeArchived
            });

            const cards = await this.apiCall(`/cards?${params}`);
            this.renderSearchResults(cards);
        } catch (error) {
            console.error('Failed to search cards:', error);
        }
    }

    renderSearchResults(cards) {
        const container = document.getElementById('searchResults');

        if (cards.length === 0) {
            container.innerHTML = '<p class="text-muted">No cards found</p>';
            return;
        }

        container.innerHTML = '';
        cards.forEach(card => {
            const resultDiv = document.createElement('div');
            resultDiv.className = 'search-result-item';
            resultDiv.innerHTML = `
                <div class="search-result-title">${this.escapeHtml(card.title)}</div>
                <div class="search-result-meta">
                    ${card.description ? this.escapeHtml(card.description) : 'No description'}
                </div>
            `;
            resultDiv.onclick = () => this.editCard(card.id);
            container.appendChild(resultDiv);
        });
    }

    // Archive
    async showArchiveModal() {
        const modal = new bootstrap.Modal(document.getElementById('archiveModal'));
        await this.loadArchivedCards();
        modal.show();
    }

    async loadArchivedCards() {
        try {
            const params = new URLSearchParams({
                board_id: this.currentBoard,
                archived: true
            });

            const cards = await this.apiCall(`/cards?${params}`);
            this.renderArchivedCards(cards);
        } catch (error) {
            console.error('Failed to load archived cards:', error);
        }
    }

    renderArchivedCards(cards) {
        const container = document.getElementById('archivedCards');

        if (cards.length === 0) {
            container.innerHTML = '<p class="text-muted">No archived cards</p>';
            return;
        }

        container.innerHTML = '';
        cards.forEach(card => {
            const cardDiv = document.createElement('div');
            cardDiv.className = 'archived-card';
            cardDiv.innerHTML = `
                <div class="archived-card-header">
                    <div class="archived-card-title">${this.escapeHtml(card.title)}</div>
                    <div class="archived-card-actions">
                        <button class="btn btn-sm btn-outline-primary" onclick="app.unarchiveCard(${card.id})">
                            <i class="bi bi-arrow-counterclockwise"></i> Restore
                        </button>
                    </div>
                </div>
                ${card.description ? `<div class="text-muted">${this.escapeHtml(card.description)}</div>` : ''}
            `;
            container.appendChild(cardDiv);
        });
    }

    async unarchiveCard(cardId) {
        try {
            await this.apiCall(`/cards/${cardId}/unarchive`, 'POST');
            await this.loadArchivedCards();
            await this.renderBoard(); // Refresh the board
            this.showAlert('Card restored successfully', 'success');
        } catch (error) {
            console.error('Failed to unarchive card:', error);
        }
    }

    // Utility Methods
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text || '';
        return div.innerHTML;
    }

    showAlert(message, type = 'info') {
        const alertDiv = document.createElement('div');
        alertDiv.className = `alert alert-${type} alert-dismissible fade show position-fixed top-0 start-50 translate-middle-x mt-3`;
        alertDiv.style.zIndex = 9999;
        alertDiv.innerHTML = `
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        `;
        document.body.appendChild(alertDiv);

        setTimeout(() => {
            alertDiv.remove();
        }, 3000);
    }
}

// Initialize the app when DOM is ready
let app;
document.addEventListener('DOMContentLoaded', () => {
    app = new KanbanApp();
});