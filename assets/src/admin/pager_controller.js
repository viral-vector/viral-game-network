import { Controller } from "@hotwired/stimulus"

export default class extends Controller {

    static values = {
        route: String,
        pages: Number,
        paged: Number
    }

    connect() {
        const pagination_list = this.element.querySelector('.pagination-list');

        // Prev & Next buttons
        if(this.pagesValue > 1){   
            if(this.pagedValue > 1){
                let prev = document.createElement('a');
                prev.setAttribute('href', `${this.routeValue}?page=${this.pagedValue - 1}`);
                prev.classList.add('pagination-previous');
                prev.innerHTML = '<i class="fas fa-caret-left"></i>';  
                this.element.insertBefore(prev, pagination_list);  
            }    
            
            if(this.pagedValue < this.pagesValue){
                let next = document.createElement('a'); 
                next.setAttribute('href', `${this.routeValue}?page=${this.pagedValue + 1}`);
                next.classList.add('pagination-next');
                next.innerHTML = '<i class="fas fa-caret-right"></i>';  
                this.element.insertBefore(next, pagination_list);
            }
        }

        pagination_list.appendChild(this.createPage(1));
        if(this.pagesValue > 1){
            // Elipses
            if(this.pagedValue > 3){
                pagination_list.appendChild(this.createDots());
            }
            // Pages
            let count = 0;
            let start = Math.max(2, this.pagedValue - 1);
            for (let i = start; i <= this.pagesValue - 1; i++) {
                pagination_list.appendChild(this.createPage(i));
                if((count++) >= 2 || i >= this.pagesValue - 1){
                    break;
                }
            }
            // Elipses
            if(this.pagesValue - this.pagedValue > 2){
                pagination_list.appendChild(this.createDots());
            }
            // Last Page
            pagination_list.appendChild(this.createPage(this.pagesValue));
        }
    }

    createPage(page) {
        const li = document.createElement('li');
        const aa = document.createElement('a');
        aa.innerHTML = `${page}`;
        aa.classList.add('pagination-link', this.pagedValue == page ? 'is-current' : 'not-current');
        aa.setAttribute('aria-label', `Page ${page}`);
        aa.setAttribute('href', `${this.routeValue}?page=${page}`);
        li.appendChild(aa);
        return li;
    }

    createDots() {
        const li = document.createElement('li');
        const sp = document.createElement('span');
        sp.innerHTML = `&mldr;`;
        sp.classList.add('.pagination-ellipsis');
        li.appendChild(sp);
        return li;
    }
}