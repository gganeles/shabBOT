function rent(prompt) {
    const monthlyUtils = prompt.match(/\d+/)
    if (monthlyUtils && parseInt(monthlyUtils[0])) {
        const monthlyExtra = parseInt(monthlyUtils)/4
        const rentDict = {
            harel:2137+monthlyExtra,
            gabe:1647+monthlyExtra,
            nico:1647+monthlyExtra,
            yehuda:1569+monthlyExtra
        }
        return Object.entries(rentDict).map(item=>item.join(': ')).join('\n')
    } else return "Please enter a valid monthly utility cost"
}

module.exports = {rent}