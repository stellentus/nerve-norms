class DataManager {
	// data is an object indexed by the drop-down's values
	// The dataUsers is expected to provide a list of objects that implement 'updateParticipant' and 'updateNorms'
	constructor(data, dataUsers) {
		this.dt = data
		this.dataUsers = dataUsers
		this.uploadCount = 0
		this.normCache = {}
		this.outlierCache = {}

		this.participants = [
			Participant.load("CA-WI20S", data),
			Participant.load("CA-AL27H", data),
			Participant.load("JP-20-1", data),
			Participant.load("JP-70-1", data),
			Participant.load("PO-00d97e84", data),
			Participant.load("PO-017182a5", data),
			Participant.load("CA Mean", data),
			Participant.load("JP Mean", data),
			Participant.load("PO Mean", data),
			Participant.load("Rat Fast Axon", data),
			Participant.load("Rat Slow Axon", data),
			Participant.load("Rat on Drugs", data),
		]

		Filter.setCallback(() => this._fetchUpdates())

		this.dropDown = document.getElementById("select-participant-dropdown")
		this.dropDown.addEventListener("change", (ev) => {
			ExVars.clearScores()
			if (this.dropDown.selectedIndex >= this.participants.length) {
				this._uploadMEM()
			} else {
				this._updateParticipant()
				this._fetchUpdates()
			}
		})
		this._updateDropDownOptions()

		this._fetchUpdates()
	}

	_fetchUpdates() {
		const lastQuery = this.queryString
		this.queryString = Filter.queryString
		const normChanged = (lastQuery != this.queryString)
		if (normChanged) {
			this._fetchNorms()
		}

		const nameChanged = (this.participantIndex != this.dropDown.selectedIndex)
		if (normChanged || nameChanged) {
			this.participantIndex = this.dropDown.selectedIndex
			const participant = this.participants[this.participantIndex]

			ExVars.clearScores()

			this._fetchOutliers(participant)
		}
	}

	_fetchNorms() {
		const query = this.queryString
		const norms = this.normCache[query]
		if (norms != null) {
			Object.values(this.dataUsers()).forEach(pl => {
				pl.updateNorms(norms)
			})
		} else {
			Fetch.Norms(query, (norms) => {
				this.normCache[query] = norms
				if (query == this.queryString) {
					// An update not has occurred since we requested this data, so update the display!
					Object.values(this.dataUsers()).forEach(pl => {
						pl.updateNorms(norms)
					})
				}
			})
		}
	}

	get _cacheString() { return this.queryString + "&id=" + this.dropDown.selectedIndex }

	_fetchOutliers(participant) {
		const cacheString = this._cacheString // Save this string because it's where we want to save the data
		const scores = this.outlierCache[cacheString]
		if (scores != null) {
			ExVars.updateScores(scores)
		} else {
			const updateAction = (scores) => {
				this.outlierCache[cacheString] = scores
				if (cacheString == this._cacheString) {
					// An update not has occurred since we requested this data, so update the display!
					ExVars.updateScores(scores)
				}
			}

			if (participant.dataIsLocal) {
				Fetch.OutliersFromName(this.queryString, participant.name, updateAction)
			} else {
				Fetch.OutliersFromJSON(this.queryString, participant.data, updateAction)
			}
		}
	}

	static get uploadOption() { return "Upload MEM..." }

	_updateDropDownOptions() {
		const selection = this.dropDown.selectedIndex

		let index = 0
		this.participants.forEach(opt => {
			this.dropDown.options[index++] = new Option(opt.name)
		})
		this.dropDown.options[index] = new Option(DataManager.uploadOption)

		if (selection >= 0) {
			this.dropDown.selectedIndex = selection
		}
	}

	_uploadMEM() {
		// This code is modified from https://stackoverflow.com/a/40971885
		var input = document.createElement('input')
		input.type = 'file'

		input.onchange = e => {
			var file = e.target.files[0]

			var reader = new FileReader()
			reader.readAsText(file, 'UTF-8')

			reader.onload = readerEvent => {
				var content = readerEvent.target.result // this is the content!
				const cacheString = this._cacheString // Save this string because it's where we want to save the data

				Fetch.MEM(this.queryString, content, convertedMem => {
					if (convertedMem.error != null) {
						console.log("Conversion error", convertedMem.error)

						this.dropDown.selectedIndex = this.participantIndex
						this.participantIndex = -1 // To force an update
						this._fetchUpdates()

						alert("The MEM could not be converted. Please email it to jbell1@ualberta.ca for troubleshooting.")
						return
					}

					this.uploadCount = this.uploadCount + 1
					const name = "Upload " + this.uploadCount + ": " + convertedMem.participant.header.name
					this.participants[this.participants.length] = new Participant(convertedMem.participant, name, false)
					this._updateDropDownOptions()

					this._updateParticipant()

					this.outlierCache[cacheString] = convertedMem.outlierScores
					if (cacheString == this._cacheString) {
						// An update has not occurred since we requested this data, so update the display!
						ExVars.updateScores(convertedMem.outlierScores)
					}
				})
			}
		}

		input.click()
	}

	_updateParticipant() {
		const data = this.participantData
		Object.values(this.dataUsers()).forEach(pl => {
			pl.updateParticipant(data)
		})

		ExVars.updateValues(data)
	}

	get norms() {
		return this.normCache[this.queryString]
	}

	get participantName() {
		return this.participants[this.dropDown.selectedIndex].name
	}

	get participantData() {
		return this.participants[this.dropDown.selectedIndex].data
	}
}
