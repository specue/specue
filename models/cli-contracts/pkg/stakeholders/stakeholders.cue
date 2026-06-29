package stakeholders

import sh "specue.io/profiles/stakeholders@v0:stakeholders"

author: sh.#Stakeholder & {
	id:    "author"
	title: "Автор контракта"
}

user: sh.#Stakeholder & {
	id:    "user"
	title: "Пользователь CLI"
}

machine: sh.#Stakeholder & {
	id:    "machine"
	title: "Скрипт или другая программа"
}
