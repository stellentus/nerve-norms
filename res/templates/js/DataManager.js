class DataManager {
	// data is an object indexed by the drop-down's values
	// The dataUsers is expected to provide a list of objects that implement 'updateParticipant' and 'updateNorms'
	constructor(data, dataUsers) {
		this.dt = data
		this.dataUsers = dataUsers
		this.uploadCount = 0

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

		Filter.setCallback(() => this.filterupdate(this.filtername))

		this.dropDown = document.getElementById("select-participant-dropdown")
		this.dropDown.addEventListener("change", (ev) => {
			this.val = ev.srcElement.value
			ExVars.clearScores()
			if (this.val == DataManager.uploadOption) {
				this._uploadMEM()
			} else {
				this.uploadData = undefined
				this._updateParticipant(this.dt[this.val])
				this.filterupdate(this.val)
			}
		})
		this._updateDropDownOptions()

		this.val = this.dropDown.value

		this.filterupdate(this.val)
	}

	filterupdate(name) {
		const lastQuery = this.filterqueryString
		this.filterqueryString = Filter.queryString
		const normChanged = (lastQuery != this.filterqueryString)
		if (normChanged) {
			Fetch.Norms(this.filterqueryString, norms => {
				this.normData = norms
				Object.values(this.dataUsers()).forEach(pl => {
					pl.updateNorms(norms)
				})
			})
		}

		if (name != undefined) {
			this.filterdata = undefined
		}
		const lastParticipant = this.filtername
		this.filtername = name
		const nameChanged = (lastParticipant != this.filtername)
		if (normChanged || nameChanged) {
			ExVars.clearScores()
			Fetch.Outliers(this.filterqueryString, this.filtername, this.filterdata)
		}
	}

	static get uploadOption() { return "Upload MEM..." }

	_updateDropDownOptions() {
		const selection = this.dropDown.selectedIndex

		let index = 0;
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
		var input = document.createElement('input');
		input.type = 'file';

		input.onchange = e => {
			var file = e.target.files[0];

			var reader = new FileReader();
			reader.readAsText(file, 'UTF-8');

			reader.onload = readerEvent => {
				var content = readerEvent.target.result; // this is the content!
				this.filterlastParticipant = undefined
				Fetch.MEM(this.filterqueryString, this.filtername, content, convertedMem => {
					this.uploadData = convertedMem.participant
					this._updateParticipant(this.uploadData)
					this.filtername = undefined
					this.filterdata = this.uploadData

					this.uploadCount = this.uploadCount + 1
					this.participants[this.participants.length] = new Participant(this.uploadData, "Upload " + this.uploadCount + ": " + this.uploadData.header.name)
					this._updateDropDownOptions()
				})
			}
		}

		input.click();
	}

	_updateParticipant(participantData) {
		Object.values(this.dataUsers()).forEach(pl => {
			pl.updateParticipant(participantData)
		})

		ExVars.updateValues(participantData)
	}

	get norms() {
		return this.normData
	}

	get participantName() {
		return this.val
	}

	get participantData() {
		if (this.uploadData != null) {
			return this.uploadData
		} else {
			return this.dt[this.val]
		}
	}
}
