class ChartFactory {
	constructor(participants) {
		this.url = "https://us-central1-nervenorms.cloudfunctions.net/"

		this.partDropDown = new DataDropDown("select-participant-dropdown", participants, (name, currentParticipant) => {
			this.participant = name

			ExVars.clearScores()
			document.getElementById("participant-name").innerHTML = name + " (loading scores...)"

			Object.values(this.plots).forEach(pl => {
				pl.updateParticipant(currentParticipant)
			})
			ExVars.updateValues(currentParticipant)
			this.updateOutliers(this.participant)
		}, ["CA-CR21S", "CA-WI20S", "Rat on Drugs", ])
		this.participant = this.partDropDown.selection

		document.getElementById("participant-name").innerHTML = this.participant + " (loading scores...)"

		const queryString = Filter.asQueryString()
		this.updateNorms(queryString)
		this.updateOutliers(this.participant, queryString)

		document.querySelector("form").addEventListener("submit", (event) => {
			ExVars.clearScores()

			const queryString = Filter.asQueryString()
			this.updateNorms(queryString)
			this.updateOutliers(this.participant, queryString)

			event.preventDefault()
		})

		this.plots = {
			"recoveryCycle": null,
			"thresholdElectrotonus": null,
			"chargeDuration": null,
			"thresholdIV": null,
			"stimulusResponse": null,
			"stimulusResponseRelative": null,
		}

		Object.keys(this.plots).forEach(key => {
			this.plots[key] = this.build(key)
			this.plots[key].draw(d3.select("#" + key + " svg"), true)
			this.plots[key].setDelayTime(Chart.fastDelay) // After initial setup, remove the delay
		})

		// Now set all excitability variables
		ExVars.updateValues(this.partDropDown.data)
	}

	/**
	 * Detect the current active responsive breakpoint in Bootstrap
	 * @returns {string}
	 * @author farside {@link https://stackoverflow.com/users/4354249/farside}
	 */
	static getResponsiveBreakpoint() {
		var envs = { xs: "d-none", sm: "d-sm-none", md: "d-md-none", lg: "d-lg-none", xl: "d-xl-none" };
		var env = "";

		var $el = $("<div>");
		$el.appendTo($("body"));

		for (var i = Object.keys(envs).length - 1; i >= 0; i--) {
			env = Object.keys(envs)[i];
			$el.addClass(envs[env]);
			if ($el.is(":hidden")) {
				break; // env detected
			}
		}
		$el.remove();
		return env;
	};

	drawModal(typeStr) {
		if (ChartFactory.getResponsiveBreakpoint() == "xs") {
			// Screen is too small to display this properly
			return
		}
		const chart = this.build(typeStr)
		chart.setDelayTime(Chart.fastDelay).setTransitionTime(Chart.fastTransition)

		document.getElementById('modal-title').innerHTML = chart.name
		d3.selectAll("#modal svg > *").remove()

		chart.draw(d3.select('#modal svg'))
		$('#modal').modal('toggle')
	}

	build(typeStr) {
		switch (typeStr) {
			case "recoveryCycle":
				return new RecoveryCycle(this.partDropDown.data, this.norms)
			case "thresholdElectrotonus":
				return new ThresholdElectrotonus(this.partDropDown.data, this.norms)
			case "chargeDuration":
				return new ChargeDuration(this.partDropDown.data, this.norms)
			case "thresholdIV":
				return new ThresholdIV(this.partDropDown.data, this.norms)
			case "stimulusResponse":
				return new StimulusResponse(this.partDropDown.data, this.norms)
			case "stimulusResponseRelative":
				return new StimulusRelative(this.partDropDown.data, this.norms)
		}
	}

	updateOutliers(name, queryString) {
		if (queryString === undefined) {
			queryString = Filter.asQueryString()
		}

		fetch(this.url + "outliers" + queryString + "&name=" + name)
			.then(response => {
				return response.json()
			})
			.then(scores => {
				this.scores = scores
				ExVars.updateScores(this.participant, scores)
			})
	}

	updateNorms(queryString) {
		if (queryString === undefined) {
			queryString = Filter.asQueryString()
		}

		fetch(this.url + "norms" + queryString)
			.then(response => {
				return response.json()
			})
			.then(norms => {
				this.norms = norms
				Object.values(this.plots).forEach(pl => {
					pl.updateNorms(norms)
				})
			})
	}
}
