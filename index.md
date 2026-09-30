## Most active GitHub users

This is a list of most active GitHub users in different countries/regions.
{% if site.data.locations.global %}
<ul class="country-list">
  <li><a href="global"><b>{{site.data.locations.global.title}}</b> (all countries/regions combined)</a></li>
</ul>
{% endif %}
<ul class="country-list">
{% assign locations = site.data.locations | sort %}
{% for loc_hash in locations %}
  {% assign location = loc_hash[1] %}
  {% unless loc_hash[0] == "global" %}
  <li><a href="{{location.page | remove: '.html'}}">{{location.title}}</a></li>
  {% endunless %}
{% endfor %}
</ul>

You can get a combined machine-readable JSON for:
<ul>
<li><a href="rank_only.json">rank-only with categories</a></li>
</ul>
A subset specific to each country/region is available on the individual page linked above.

### Badges

Badges are also available, which you can include on your profile pages. Simply include the following markdown for users:
```markdown
[![committers.top badge](https://user-badge.committers.top/REGION/USERNAME.svg)](https://user-badge.committers.top/REGION/USERNAME)
```
For organizations, you need to use a slightly different markup:
```markdown
[![committers.top badge](https://org-badge.committers.top/REGION/ORGNAME.svg)](https://org-badge.committers.top/REGION/ORGNAME)
```
Use `global` as the `REGION` for the combined ranking of all countries/regions. In case you aren't currently ranked for a given region, you'll simply receive an "unranked" badge.
