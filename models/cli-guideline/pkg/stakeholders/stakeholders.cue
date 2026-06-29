package stakeholders

import sh "specue.io/profiles/stakeholders@v0:stakeholders"

developer: sh.#Stakeholder & {
	id:    "developer"
	title: "Разработчик CLI"
}

user: sh.#Stakeholder & {
	id:    "user"
	title: "Пользователь CLI"
}
